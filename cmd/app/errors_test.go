package main

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestLogError(t *testing.T) {
	err := errors.New("test error")
	req := httptest.NewRequest("GET", "/test", nil)

	var buf bytes.Buffer
	app := application{
		logger: log.New(&buf),
	}

	app.logError(req, err)

	got := buf.String()
	got = strings.TrimSuffix(got, "\n")
	if !strings.Contains(got, "test error") {
		t.Errorf("logError() = %v, expected to contain 'test error'", got)
	}
	if !strings.Contains(got, "GET") {
		t.Errorf("logError() = %v, expected to contain 'GET'", got)
	}
	if !strings.Contains(got, "/test") {
		t.Errorf("logError() = %v, expected to contain '/test'", got)
	}
}

func TestErrResponse(t *testing.T) {}
