package main

import (
	"reflect"
	"testing"
)

func TestValidateUrl(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "empty",
			input:       "",
			expectError: true,
		},
		{
			name:     "absolute url",
			input:    "https://example.com/pages/",
			expected: "https://example.com/pages/",
		},
		{
			name:        "non absolute url",
			input:       "/pages/",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			urlStruct, err := validateUrl(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf(
						"expected error for input %v, but got none",
						tt.input,
					)
				}
				return
			}
			if err != nil {
				t.Errorf(
					"unexpected error: %v, for input %v",
					err,
					tt.input,
				)
			}

			got := urlStruct.String()

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf(
					"input: %v, expected: %v, got: %v",
					tt.input,
					tt.expected,
					got,
				)
			}
		})
	}
}
