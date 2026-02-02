package main

import (
	"io"
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

func (app *application) getUrlFromHTML(htmlReader io.Reader, baseUrl string) ([]string, error) {
	urls := []string{}
	baseUrlStruct, err := url.ParseRequestURI(baseUrl)
	if err != nil {
		return nil, err
	}
	htmlNode, err := html.Parse(htmlReader)
	if err != nil {
		return nil, err
	}
	for node := range htmlNode.Descendants() {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, att := range node.Attr {
				if att.Key == "href" {
					u, err := url.ParseRequestURI(att.Val)
					if err != nil {
						continue
					}

					u = baseUrlStruct.ResolveReference(u)
					urls = append(urls, u.String())
				}
			}
		}
	}
	return urls, nil
}
