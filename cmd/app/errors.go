package main

import (
	"fmt"
	"net/http"
)

func (app *application) serverErrResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	msg := "the server encountered a problem and could not process your request"
	app.errResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *application) errResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	env := envelope{"error": msg}
	if err := app.writeJSON(w, env, status); err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err, "method", method, "uri", uri)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "the request resource could not be found"
	app.errResponse(w, r, http.StatusNotFound, msg)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("the %s method is not supported for this response", r.Method)
	app.errResponse(w, r, http.StatusMethodNotAllowed, msg)
}
