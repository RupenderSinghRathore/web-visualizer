package main

import (
	"errors"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, envelope{"status": "ok"}, http.StatusOK)
}

var (
	ErrInvalidUrl     = errors.New("invalid url")
	ErrNonAbsoluteUrl = errors.New("non absolute url")
)

func (app *application) fetchGraphHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Url string `json:"url"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.errResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	urlStruct, err := validateUrl(input.Url)
	if err != nil {
		app.errResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	graph := app.crawlUrl(urlStruct)
	env := envelope{"graph": graph}
	if err := app.writeJSON(w, env, http.StatusOK); err != nil {
		app.errResponse(w, r, http.StatusBadRequest, err.Error())
	}
}
