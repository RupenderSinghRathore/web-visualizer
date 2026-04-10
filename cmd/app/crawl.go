package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"

	"RupenderSinghRathore/web-visualizer/internal/data"

	"golang.org/x/net/html"
)

type crawlResult struct {
	url    string
	links  []string
	status int
	err    error
	depth  int
}

var (
	ErrStatusCode  = errors.New("error status code")
	ErrContentType = errors.New("error content-type")
)

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

func (app *application) fetch(urlStrt urlStrDepth) *crawlResult {
	result := crawlResult{url: urlStrt.url}

	req, err := http.NewRequest("GET", urlStrt.url, nil)
	if err != nil {
		result.status = http.StatusNotFound
		result.err = err
		return &result
	}
	req.Header.Set("User-Agent", "WebVisualizer/1.0")
	req.Header.Set("Accept", "text/html")

	res, err := app.client.Do(req)
	if err != nil {
		result.err = err
		return &result
	}
	defer res.Body.Close()

	status := res.StatusCode
	result.status = status
	if status > 399 {
		// result.err = fmt.Errorf("%s: %d status code", urlStr, res.StatusCode)
		result.err = ErrStatusCode
		return &result
	}
	if !isHTML(res.Header.Get("content-type")) {
		result.err = ErrContentType
		// result.err = fmt.Errorf("%s: %s content-type", urlStr, res.Header.Get("content-type"))
		return &result
	}

	finalUrl := res.Request.URL

	linksMap, err := app.extractLinksFromBody(res.Body, finalUrl)

	links := make([]string, 0, len(linksMap))
	for link := range linksMap {
		if link != urlStrt.url && link != finalUrl.String() {
			links = append(links, link)
		}
	}
	slices.Sort(links)

	result.links = links
	result.depth = urlStrt.depth
	return &result
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
						link = strings.TrimSpace(link)
						resolvedLink, err := url.Parse(link)
						if err != nil {
							app.logger.Error(err)
							continue
						}
						resolvedLink = baseUrl.ResolveReference(resolvedLink)
						if resolvedLink.Hostname() != baseUrl.Hostname() {
							continue
						}
						// if !strings.HasPrefix(resolvedLink.Path, baseUrl.Path) {
						// 	continue
						// }
						link = resolvedLink.String()
						links[link] = struct{}{}
					}
				}
			}
		}
	}
}

type urlStrDepth struct {
	url   string
	depth int
}

func (app *application) crawlUrl(urlB *url.URL) data.Graph {
	normalizedBase := app.getPath(urlB)

	graph := data.Graph{}
	seen := make(map[string]bool)

	results := make(chan *crawlResult)
	workQueue := make(chan urlStrDepth, 100)

	seen[normalizedBase] = true
	workQueue <- urlStrDepth{urlB.String(), 1}

	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	wg := sync.WaitGroup{}
	for i := 0; i < app.config.crawl.maxGoroutine; i++ {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case urlStrt := <-workQueue:
					result := app.fetch(urlStrt)
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

		switch {
		case result.err != nil:
			switch {
			case result.status == http.StatusTooManyRequests:
				fmt.Fprintf(os.Stderr, EraseLineANSI)
				app.logger.Error(result.err)
				cancle()
			case errors.Is(result.err, ErrStatusCode) || errors.Is(result.err, ErrContentType):
			default:
				fmt.Fprintf(os.Stderr, EraseLineANSI)
				app.logger.Error(result.err)
			}
		case result.depth > app.config.crawl.maxDepth:
			fmt.Fprintf(os.Stderr, EraseLineANSI)
			app.logger.Error("max depth reached.")
		}

		urlStruct, err := url.Parse(result.url)
		if err != nil {
			app.logger.Error(err)
			continue
		}
		currPath := app.getPath(urlStruct)

		edge := &data.Edge{Visited: 1, Depth: result.depth, Status: result.status, Links: []string{}}
		graph[currPath] = edge

		for _, link := range result.links {

			urlStruct, err := url.Parse(link)
			if err != nil {
				app.logger.Error(err)
				continue
			}
			childPath := app.getPath(urlStruct)

			edge.Links = append(edge.Links, childPath)
			if seen[childPath] {
				if e, ok := graph[childPath]; ok {
					e.Visited++
				}
				continue
			}
			seen[childPath] = true

			fetching++
			go func(l urlStrDepth) {
				select {
				case workQueue <- l:
				case <-ctx.Done():
					return
				}
			}(urlStrDepth{link, edge.Depth + 1})

		}
	}

	cancle()
	wg.Wait()

	return graph
}
