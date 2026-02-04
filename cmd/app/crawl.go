package main

import (
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

func (app *application) fetchLinks(urlStr string) ([]string, error) {
	res, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 399 {
		return nil, fmt.Errorf("%d status code", res.StatusCode)
	}
	if !isHTML(res.Header.Get("content-type")) {
		return nil, fmt.Errorf("%s: %s content-type", urlStr, res.Header.Get("content-type"))
	}

	return app.extractLinksFromBody(res.Body, urlStr)
}

// needs testing
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

func (app *application) crawlPage(baseUrl string) (map[string]struct{}, error) {
	pages := make(map[string]struct{})
	urlB, err := url.ParseRequestURI(baseUrl)
	if err != nil {
		return nil, err
	}
	normalizedUrl, err := app.normalizeUrl(baseUrl)
	if err != nil {
		return nil, err
	}
	pages[normalizedUrl] = struct{}{}

	result := make(chan []string)

	workQueue := make(chan string, 100)
	workQueue <- urlB.String()

	fetching := 1
	for i := 0; i < app.currencyLimit; i++ {
		go func() {
			for urlStr := range workQueue {
				links, err := app.fetchLinks(urlStr)
				if err != nil {
					app.logger.Error(err)
				}
				result <- links
			}
		}()
	}

	for fetching > 0 {
		newLinks := <-result
		fetching--

		for _, link := range newLinks {

			normalizedLink, err := app.normalizeUrl(link)
			if err != nil {
				continue
			}

			if _, ok := pages[normalizedLink]; !ok {
				pages[normalizedLink] = struct{}{}

				fetching++
				go func(l string) { workQueue <- l }(link)
			}

		}
	}

	close(workQueue)
	close(result)

	return pages, nil
}
