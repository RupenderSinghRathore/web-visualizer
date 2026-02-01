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
			name:        "empty url",
			input:       "",
			expected:    "",
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
