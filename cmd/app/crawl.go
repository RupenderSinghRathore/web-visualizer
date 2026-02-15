package main

import (
	"RupenderSinghRathore/web-visualizer/internal/data"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"

	"golang.org/x/net/html"
)

type crawlResult struct {
	parent string
	links  []string
	status int
	err    error
}

func (app *application) getPath(urlStruct *url.URL) string {
	path := urlStruct.Path

	if len(path) == 0 {
		path = "/"
	}

	if path != "/" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

func (app *application) fetch(urlStr string) *crawlResult {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return &crawlResult{parent: urlStr, err: err}
	}
	req.Header.Set("User-Agent", "WebVisualizer/1.0")
	req.Header.Set("Accept", "text/html")

	res, err := app.client.Do(req)
	if err != nil {
		return &crawlResult{parent: urlStr, err: err}
	}
	defer res.Body.Close()

	status := res.StatusCode
	if status > 399 {
		return &crawlResult{parent: urlStr, status: status, err: fmt.Errorf(
			"%s: %d status code",
			urlStr,
			res.StatusCode,
		)}
	}
	if !isHTML(res.Header.Get("content-type")) {
		return &crawlResult{parent: urlStr, status: status, err: fmt.Errorf(
			"%s: %s content-type",
			urlStr,
			res.Header.Get("content-type"),
		)}
		// return &crawlResult{parent: urlStr, status: status, err: nil}
	}

	finalUrl := res.Request.URL

	linksMap, err := app.extractLinksFromBody(res.Body, finalUrl)

	links := make([]string, 0, len(linksMap))
	for link := range linksMap {
		if link != urlStr && link != finalUrl.String() {
			links = append(links, link)
		}
	}
	slices.Sort(links)

	return &crawlResult{
		parent: urlStr,
		status: status,
		links:  links,
	}
}

func (app *application) extractLinksFromBody(
	body io.Reader,
	baseUrl *url.URL,
) (map[string]struct{}, error) {
	links := map[string]struct{}{}
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
						resolvedLink, err := url.Parse(link)
						if err != nil {
							app.logger.Error(err)
							continue
						}
						resolvedLink = baseUrl.ResolveReference(resolvedLink)
						if resolvedLink.Hostname() != baseUrl.Hostname() {
							continue
						}
						link = resolvedLink.String()
						links[link] = struct{}{}
					}
				}
			}
		}
	}
}

func (app *application) crawlPage(urlB *url.URL) data.Graph {
	normalizedBase := app.getPath(urlB)

	graph := data.Graph{}
	seen := make(map[string]bool)

	results := make(chan *crawlResult)
	workQueue := make(chan string, 100)

	seen[normalizedBase] = true
	workQueue <- urlB.String()

	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	wg := sync.WaitGroup{}
	for i := 0; i < app.config.crawl.maxGoroutine; i++ {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case urlStr := <-workQueue:
					result := app.fetch(urlStr)
					select {
					case results <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		})

	}

	fetching := 1
out:
	for fetching > 0 && len(graph) < app.config.crawl.maxPages {
		var result *crawlResult
		select {
		case result = <-results:
		case <-ctx.Done():
			break out
		}
		fetching--

		if result.err != nil {
			fmt.Fprintf(os.Stderr, EraseLineANSI)
			app.logger.Error(result.err)
			if result.status == http.StatusTooManyRequests {
				cancle()
			}
		}

		urlStruct, err := url.Parse(result.parent)
		normParent := app.getPath(urlStruct)
		if err != nil {
			app.logger.Error(err)
			continue
		}

		edge := &data.Edge{Visited: 1, Status: result.status, Links: []string{}}
		graph[normParent] = edge

		for _, link := range result.links {

			urlStruct, err := url.Parse(link)
			normLink := app.getPath(urlStruct)
			if err != nil {
				app.logger.Error(err)
				continue
			}

			edge.Links = append(edge.Links, normLink)
			if seen[normLink] {
				if e, ok := graph[normLink]; ok {
					e.Visited++
				}
				continue
			}
			seen[normLink] = true

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

	cancle()
	wg.Wait()

	return graph
}
