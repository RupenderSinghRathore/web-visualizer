package main

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func (app *application) normalizeUrl(urlS string) (string, error) {
	u, err := url.Parse(urlS)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	if host != "" {
		host = strings.ToLower(host)
	}
	path := u.Path
	if path != "" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return host + path, nil
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
	queue := []string{urlB.String()}

	for len(queue) != 0 {
		currUrl := queue[0]
		queue = queue[1:]

		res, err := http.Get(currUrl)
		if err != nil {
			app.logger.Error(err)
			continue
		}
		defer res.Body.Close()

		if res.StatusCode > 399 || !isHTML(res.Header.Get("content-type")) {
			continue
		}

		htmlNode, err := html.Parse(res.Body)
		if err != nil {
			app.logger.Error(err)
			continue
		}

		for node := range htmlNode.Descendants() {
			if node.Type == html.ElementNode && node.Data == "a" {
				for _, att := range node.Attr {
					if att.Key == "href" {
						nextUrl, err := url.ParseRequestURI(att.Val)
						if err != nil {
							app.logger.Error(err)
							continue
						}

						nextUrl = urlB.ResolveReference(nextUrl)

						if nextUrl.Hostname() != urlB.Hostname() {
							continue
						}

						normalizedNextUrl, err := app.normalizeUrl(nextUrl.String())
						if err != nil {
							app.logger.Error(err)
							continue
						}
						if _, ok := pages[normalizedNextUrl]; ok {
							continue
						}
						pages[normalizedNextUrl] = struct{}{}

						queue = append(queue, nextUrl.String())
					}
				}
			}
		}
	}
	return pages, nil
}
