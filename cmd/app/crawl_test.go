package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeUrl(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "empty url",
			input:    "",
			expectError: true,
		},
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
			got, err := app.getPath(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %v, but got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for input %v, but got: %v", tt.input, err)
			}
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

			got, err := app.extractLinksFromBody(reader, tt.baseUrl)

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
