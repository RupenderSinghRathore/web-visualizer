package main

import (
	"RupenderSinghRathore/web-visualizer/internal/data"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "base url",
			input:    "http://google.com",
			expected: "/",
		},
		{
			name:     "standard http url",
			input:    "http://google.com/pages",
			expected: "/pages",
		},
		{
			name:     "standard https url",
			input:    "https://google.com/pages",
			expected: "/pages",
		},
		{
			name:     "url with query parameters",
			input:    "https://google.com/pages?page=3",
			expected: "/pages",
		},
		{
			name:     "url with trailing /",
			input:    "https://google.com/pages/",
			expected: "/pages",
		},
		{
			name:     "url with capitals",
			input:    "https://GOOgle.com/pages",
			expected: "/pages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := application{}
			urlStruct, err := url.Parse(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v, for input %v", tt.input, err)
			}

			got := app.getPath(urlStruct)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf(
					"test %v failed: input: %v, expected: %v, got: %v",
					tt.name,
					tt.input,
					tt.expected,
					got,
				)
			}
		})
	}
}

func TestExtractLinksFromBody(t *testing.T) {
	tests := []struct {
		name        string
		htmlBody    string
		baseUrl     string
		expected    map[string]struct{}
		expectError bool
	}{
		{
			name:     "empty html",
			htmlBody: "",
			baseUrl:  "https://google.com/pages",
			expected: map[string]struct{}{},
		},
		{
			name:     "no url",
			htmlBody: `<html><body></body></html>`,
			baseUrl:  "https://google.com/pages",
			expected: map[string]struct{}{},
		},
		{
			name: "single url",
			htmlBody: `
			<html>
			  <body>
				<a href="https://www.google.com/">Visit Google.com</a>
			  </body>
			</html>
			`,
			baseUrl:  "https://www.google.com",
			expected: map[string]struct{}{"https://www.google.com/": {}},
		},
		{
			name: "multiple urls",
			htmlBody: `
			<html>
			  <body>
				<a href="https://www.google.com/">Visit Google.com</a>
				<a href="https://www.google.com/pages">Visit Google.com</a>
				<a href="https://www.moogle.com/">Visit Google.com</a>
				<a href="https://www.doodle.com/">Visit Google.com</a>
			  </body>
			</html>
			`,
			baseUrl: "https://www.google.com",
			expected: map[string]struct{}{
				"https://www.google.com/":      {},
				"https://www.google.com/pages": {},
			},
		},
		{
			name: "relative url",
			htmlBody: `
			<html>
			  <body>
				<a href="/pages/">Visit Google.com</a>
			  </body>
			</html>
			`,
			baseUrl: "https://google.com",
			expected: map[string]struct{}{
				"https://google.com/pages/": {},
			},
		},
		{
			name: "both urls",
			htmlBody: `
			<html>
			  <body>
				<a href="https://www.google.com/">Visit Google.com</a>
				<a href="/pages/">Visit Google.com</a>
			  </body>
			</html>
			`,
			baseUrl: "https://www.google.com",
			expected: map[string]struct{}{
				"https://www.google.com/":       {},
				"https://www.google.com/pages/": {},
			},
		},
		{
			name: "relative url(non-absolute path)",
			htmlBody: `
			<html>
			  <body>
				<a href="foo.html">Visit Google.com</a>
			  </body>
			</html>
			`,
			baseUrl: "https://www.google.com",
			expected: map[string]struct{}{
				"https://www.google.com/foo.html": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := application{}
			reader := strings.NewReader(tt.htmlBody)

			urlStruct, err := url.Parse(tt.baseUrl)
			if err != nil {
				t.Errorf(
					"unexpected error for input { %v, %v }, but got: %v",
					tt.htmlBody,
					tt.baseUrl,
					err,
				)
			}

			got, err := app.extractLinksFromBody(reader, urlStruct)

			if tt.expectError {
				if err == nil {
					t.Errorf(
						"expected error for input { %v, %v }, but got none",
						tt.htmlBody,
						tt.baseUrl,
					)
				}
				return
			}

			if err != nil {
				t.Errorf(
					"unexpected error for input { %v, %v }, but got: %v",
					tt.htmlBody,
					tt.baseUrl,
					err,
				)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf(
					"test %v failed: input: { %v, %v }, expected: %v, got: %v",
					tt.name,
					tt.htmlBody,
					tt.baseUrl,
					tt.expected,
					got,
				)
			}
		})
	}
}

func TestCrawlUrl(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `<a href="/about"></a>
					   <a href="/contact"></a>
					   <a href="/details"></a>`)
		case "/about":
			fmt.Fprint(w, `<a href="/"></a>`)
		case "/contact":
			fmt.Fprint(w, `<a href="/"></a>`)
		case "/details":
			fmt.Fprint(w, `<a href="/"></a>`)
		}
	}))
	defer s.Close()

	tests := []struct {
		name     string
		input    string
		expected data.Graph
	}{
		{
			name:  "general page",
			input: s.URL,
			expected: data.Graph{
				"/": &data.Edge{
					Visited: 4,
					Status:  200,
					Links:   []string{"/about", "/contact", "/details"},
				},
				"/about": &data.Edge{
					Visited: 1,
					Status:  200,
					Links:   []string{"/"},
				},
				"/contact": &data.Edge{
					Visited: 1,
					Status:  200,
					Links:   []string{"/"},
				},
				"/details": &data.Edge{
					Visited: 1,
					Status:  200,
					Links:   []string{"/"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg confugration
			cfg.crawl.maxGoroutine = 20
			cfg.crawl.maxPages = 1000
			app := application{
				config: &cfg,
				client: &http.Client{
					Timeout: 10 * time.Second,
				},
			}

			urlStruct, err := url.ParseRequestURI(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v, for input: %v", err, tt.input)
			}
			got := app.crawlUrl(urlStruct)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("input: %v, expected: %v, got: %v", tt.input, tt.expected, got)
			}
		})
	}
}
