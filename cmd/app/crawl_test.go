package main

import (
	"reflect"
	"testing"
)

func TestNormalizeUrl(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		got         string
		expected    string
		expectError bool
	}{
		{
			name:     "empty url",
			input:    "",
			expected: "",
		},
		{
			name:     "standard http url",
			input:    "http://google.com/pages",
			expected: "google.com/pages",
		},
		{
			name:     "standard https url",
			input:    "https://google.com/pages",
			expected: "google.com/pages",
		},
		{
			name:     "url with query parameters",
			input:    "https://google.com/pages?page=3",
			expected: "google.com/pages",
		},
		{
			name:     "url with trailing /",
			input:    "https://google.com/pages/",
			expected: "google.com/pages",
		},
		{
			name:     "url with capitals",
			input:    "https://GOOgle.com/pages",
			expected: "google.com/pages",
		},
	}

	for _, tt := range tests {
		var err error
		t.Run(t.Name(), func(t *testing.T) {
			app := application{}
			tt.got, err = app.normalizeUrl(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %v, but got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for input %v, but got: %v", tt.input, err)
			}
			if !reflect.DeepEqual(tt.got, tt.expected) {
				t.Errorf(
					"test %v failed: input: %v, expected: %v, got: %v",
					tt.name,
					tt.input,
					tt.expected,
					tt.got,
				)
			}
		})
	}
}

// func TestGetUrlFromHTML(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		htmlReader  io.Reader
// 		baseUrl     string
// 		got         []string
// 		expected    []string
// 		expectError bool
// 	}{
// 		{
// 			name:       "empty html",
// 			htmlReader: strings.NewReader(""),
// 			baseUrl:    "https://google.com/pages",
// 			expected:   []string{},
// 		},
// 		{
// 			name:       "no url from html",
// 			htmlReader: strings.NewReader(`<html><body></body></html>`),
// 			baseUrl:    "https://google.com/pages",
// 			expected:   []string{},
// 		},
// 		{
// 			name: "single url from html",
// 			htmlReader: strings.NewReader(`
// 			<html>
// 			  <body>
// 				<a href="https://www.google.com/">Visit Google.com</a>
// 			  </body>
// 			</html>
// 			`),
// 			baseUrl:  "https://www.google.com",
// 			expected: []string{"https://www.google.com/"},
// 		},
// 		{
// 			name: "multiple urls from html",
// 			htmlReader: strings.NewReader(`
// 			<html>
// 			  <body>
// 				<a href="https://www.google.com/">Visit Google.com</a>
// 				<a href="https://www.moogle.com/">Visit Google.com</a>
// 				<a href="https://www.doodle.com/">Visit Google.com</a>
// 			  </body>
// 			</html>
// 			`),
// 			baseUrl: "https://www.google.com",
// 			expected: []string{
// 				"https://www.google.com/",
// 				"https://www.moogle.com/",
// 				"https://www.doodle.com/",
// 			},
// 		},
// 		{
// 			name: "relative url from html",
// 			htmlReader: strings.NewReader(`
// 			<html>
// 			  <body>
// 				<a href="/pages/">Visit Google.com</a>
// 			  </body>
// 			</html>
// 			`),
// 			baseUrl:  "https://google.com",
// 			expected: []string{"https://google.com/pages/"},
// 		},
// 		{
// 			name: "both urls from html",
// 			htmlReader: strings.NewReader(`
// 			<html>
// 			  <body>
// 				<a href="https://www.google.com/">Visit Google.com</a>
// 				<a href="/pages/">Visit Google.com</a>
// 			  </body>
// 			</html>
// 			`),
// 			baseUrl: "https://www.google.com",
// 			expected: []string{
// 				"https://www.google.com/",
// 				"https://www.google.com/pages/",
// 			},
// 		},
// 		{
// 			name: "invalid url from html",
// 			htmlReader: strings.NewReader(`
// 			<html>
// 			  <body>
// 				<a href="invalid">Visit Google.com</a>
// 			  </body>
// 			</html>
// 			`),
// 			baseUrl:  "https://www.google.com",
// 			expected: []string{},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(t.Name(), func(t *testing.T) {
// 			app := application{}
// 			htmlByteBody, err := io.ReadAll(tt.htmlReader)
// 			tt.htmlReader = bytes.NewReader(htmlByteBody)
// 			htmlBody := string(htmlByteBody)
// 			if err != nil {
// 				t.Error(err.Error())
// 			}
// 			tt.got, err = app.getUrlFromHTML(tt.htmlReader, tt.baseUrl)
// 			if tt.expectError {
// 				if err == nil {
// 					t.Errorf(
// 						"expected error for input { %v, %v }, but got none",
// 						htmlBody,
// 						tt.baseUrl,
// 					)
// 				}
// 				return
// 			}
// 			if err != nil {
// 				t.Errorf(
// 					"unexpected error for input { %v, %v }, but got: %v",
// 					htmlBody,
// 					tt.baseUrl,
// 					err,
// 				)
// 			}
// 			if !reflect.DeepEqual(tt.got, tt.expected) {
// 				t.Errorf(
// 					"test %v failed: input: { %v, %v }, expected: %v, got: %v",
// 					tt.name,
// 					htmlBody,
// 					tt.baseUrl,
// 					tt.expected,
// 					tt.got,
// 				)
// 			}
// 		})
// 	}
// }
