package main

import (
	"net/url"
	"strings"
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
