package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	EraseLineANSI = "\r\033[K"
	MegaByte      = 1_048_576
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, data envelope, status int) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := MegaByte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf(
				"body contains badly-formed JSON (at character %d)",
				syntaxError.Offset,
			)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf(
					"body contains incorrect JSON type for field %q",
					unmarshalTypeError.Field,
				)
			}
			return fmt.Errorf(
				"body contains incorrect JSON type (at character %d)",
				unmarshalTypeError.Offset,
			)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		default:
			return err
		}
	}
	return nil
}

func (app *application) crashErr(err error) {
	app.logger.Error(err)
	os.Exit(1)
}
func (app *application) spinningAnimation(ch <-chan struct{}, wg *sync.WaitGroup) {
	wg.Done()
	spinner := `-\|/`
	n := len(spinner)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-ch:
			return
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "\r%c", spinner[i])
			i = (i + 1) % n
		}
	}
}
func isHTML(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return mediaType == "text/html"
}

func validateUrl(u string) (*url.URL, error) {
	urlStruct, err := url.ParseRequestURI(u)
	switch {
	case err != nil:
		err = ErrInvalidUrl
	case !urlStruct.IsAbs():
		err = ErrNonAbsoluteUrl
	}
	return urlStruct, err
}

func linkedText(u string, status, visited int, base string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s(%d, %d)\x1b]8;;\x1b\\", base+u, u, status, visited)
}
func webLinkedTag(u string, status, visited int, base string) string {
	return fmt.Sprintf("<a href='%s' >%s(%d, %d)</a>", base+u, u, status, visited)
}
