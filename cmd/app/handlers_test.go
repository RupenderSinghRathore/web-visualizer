package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
)

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "fuckyou", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	w := httptest.NewRecorder()

	app := &application{}
	app.healthcheckHandler(w, req)

	expected := "<h1>Heavenly Demon God Domain</h1>"
	got := w.Body.String()

	if got != expected {
		t.Errorf("expected: %s, got: %s", expected, got)
	}
}

func TestFetchGraphHandler(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected envelope
	}{
		{
			name:     "invalid url",
			input:    "pagesdeeznuts",
			expected: envelope{"error": ErrInvalidUrl.Error()},
		},
		{
			name:     "non-absolute url",
			input:    "/pages/some",
			expected: envelope{"error": ErrNonAbsoluteUrl.Error()},
		},
		{
			name:  "valid url",
			input: "kkks://valid.com",
			expected: envelope{
				"graph": map[string]any{
					"/": map[string]any{
						"visited": float64(1),
						"links":   []any{},
						"status":  float64(0),
					},
				},
			},
		},
	}

	var cfg confugration
	cfg.crawl.maxGoroutine = 20
	cfg.crawl.maxPages = 1000
	app := application{
		config: &cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: log.New(io.Discard),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(fmt.Sprintf("{\"url\":\"%s\"}", tt.input))
			req, err := http.NewRequest("GET", "fuckyou", body)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			w := httptest.NewRecorder()

			app.fetchGraphHandler(w, req)

			got := envelope{}
			if err := json.NewDecoder(w.Result().Body).Decode(&got); err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("input: %v, expected: %+v, got: %+v", tt.input, tt.expected, got)
			}
		})
	}
}
