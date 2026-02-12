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
	return path, nil
}

func (app *application) fetchLinks(urlStr string) (urlPacket, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return urlPacket{}, err
	}
	req.Header.Set("User-Agent", "WebVisualizer/1.0")
	req.Header.Set("Accept", "text/html")

	res, err := app.client.Do(req)
	if err != nil {
		return urlPacket{}, err
	}
	defer res.Body.Close()

	status := res.StatusCode
	if status > 399 {
		return urlPacket{status: status}, fmt.Errorf("%s: %d status code", urlStr, res.StatusCode)
	}
	if !isHTML(res.Header.Get("content-type")) {
		// return urlPacket{status: status}, fmt.Errorf(
		// 	"%s: %s content-type",
		// 	urlStr,
		// 	res.Header.Get("content-type"),
		// )
		return urlPacket{status: status}, nil
	}

	finalUrl := res.Request.URL.String()

	links, err := app.extractLinksFromBody(res.Body, finalUrl)
	return urlPacket{finalUrl, status, links}, err
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

type urlPacket struct {
	parent string
	status int
	links  []string
}

func (app *application) crawlPage(urlB *url.URL) (data.Graph, error) {
	normalizedBase, err := app.normalizeUrl(urlB.String())
	if err != nil {
		return nil, err
	}

	graph := data.Graph{}
	seen := make(map[string]bool)
	seen[normalizedBase] = true

	result := make(chan urlPacket)
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
					packet, err := app.fetchLinks(urlStr)
					if err != nil {
						app.logger.Error(err)
						if packet.status == http.StatusTooManyRequests {
							cancle()
						}
					}

					select {
					case result <- packet:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	fetching := 1
	for fetching > 0 && len(graph) < app.config.crawl.maxPages {
		var packet urlPacket
		select {
		case packet = <-result:
		case <-ctx.Done():
			return graph, nil
		}

		currLink := packet.parent
		newLinks := packet.links
		status := packet.status

		fetching--

		normalizedCurrLink, err := app.normalizeUrl(currLink)
		if err != nil {
			app.logger.Error(err)
			continue
		}

		edge := &data.Edge{Visited: 1, Status: status, Links: map[string]struct{}{}}
		graph[normalizedCurrLink] = edge

		for _, link := range newLinks {

			normalizedLink, err := app.normalizeUrl(link)
			if err != nil {
				app.logger.Error(err)
				continue
			}

			edge.Links[normalizedLink] = struct{}{}
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
