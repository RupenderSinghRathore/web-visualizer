package main

import (
	"RupenderSinghRathore/web-visualizer/internal/data"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func (app *application) normalizeUrl(urlStr string) (string, error) {
	urlStruct, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	host := urlStruct.Hostname()
	if host != "" {
		host = strings.ToLower(host)
	}
	path := urlStruct.Path
	if path != "" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return host + path, nil
}

func (app *application) fetchLinks(urlStr string) (string, []string, int, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", nil, 0, err
	}
	req.Header.Set("User-Agent", "WebVisualizer/1.0")
	req.Header.Set("Accept", "text/html")

	res, err := app.client.Do(req)
	if err != nil {
		return "", nil, 0, err
	}
	defer res.Body.Close()

	status := res.StatusCode
	if status > 399 {
		return "", nil, status, fmt.Errorf("%s: %d status code", urlStr, res.StatusCode)
	}
	if !isHTML(res.Header.Get("content-type")) {
		return "", nil, status, fmt.Errorf(
			"%s: %s content-type",
			urlStr,
			res.Header.Get("content-type"),
		)
	}

	finalUrl := res.Request.URL.String()

	links, err := app.extractLinksFromBody(res.Body, finalUrl)
	return finalUrl, links, status, err
}

func (app *application) extractLinksFromBody(body io.Reader, urlStr string) ([]string, error) {
	baseUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	var links []string
	tokenizer := html.NewTokenizer(body)

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			return links, nil
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link := attr.Val
						linkStruct, err := url.Parse(link)
						if err != nil {
							app.logger.Error(err)
							continue
						}
						linkStruct = baseUrl.ResolveReference(linkStruct)
						if linkStruct.Hostname() != baseUrl.Hostname() {
							continue
						}
						link = linkStruct.String()
						links = append(links, link)
					}
				}
			}
		}
	}
}

type send struct {
	parent string
	status int
	links  []string
}

func (app *application) crawlPage(baseUrl string) (data.Graph, error) {
	urlB, err := url.ParseRequestURI(baseUrl)
	if err != nil {
		return nil, err
	}

	normalizedBase, err := app.normalizeUrl(urlB.String())
	if err != nil {
		return nil, err
	}

	graph := data.Graph{}
	seen := make(map[string]bool)
	seen[normalizedBase] = true

	result := make(chan send)
	workQueue := make(chan string, 100)
	workQueue <- urlB.String()

	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	for i := 0; i < app.config.crawl.maxGoroutine; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case urlStr := <-workQueue:
					finalUrl, links, status, err := app.fetchLinks(urlStr)
					if err != nil {
						app.logger.Error(err)
						if status == http.StatusTooManyRequests {
							cancle()
						}
					}

					select {
					case result <- send{finalUrl, status, links}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	fetching := 1
	for fetching > 0 && len(graph) < app.config.crawl.maxPages {
		var d send
		select {
		case d = <-result:
		case <-ctx.Done():
			return graph, nil
		}

		currLink := d.parent
		newLinks := d.links
		status := d.status

		fetching--

		normalizedCurrLink, err := app.normalizeUrl(currLink)
		if err != nil {
			app.logger.Error(err)
			continue
		}

		edge := &data.Edge{Visited: 1, Status: status}
		graph[normalizedCurrLink] = edge

		for _, link := range newLinks {

			normalizedLink, err := app.normalizeUrl(link)
			if err != nil {
				app.logger.Error(err)
				continue
			}

			edge.Links = append(edge.Links, normalizedLink)
			if seen[normalizedLink] {
				if e, ok := graph[normalizedLink]; ok {
					e.Visited++
				}
				continue
			}
			seen[normalizedLink] = true

			fetching++
			go func(l string) {
				select {
				case workQueue <- l:
				case <-ctx.Done():
					return
				}
			}(link)

		}
	}

	return graph, nil
}
