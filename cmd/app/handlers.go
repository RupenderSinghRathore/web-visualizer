package main

import (
	"fmt"
	"net/http"
	"net/url"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Heavenly Demon God Domain</h1>")
}

func (app *application) getGraph(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Url string `json:"url"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.errResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	urlStruct, err := url.ParseRequestURI(input.Url)
	if err != nil {
		app.errResponse(w, r, http.StatusBadRequest, "invalid url")
		return
	}
	if !urlStruct.IsAbs() {
		app.errResponse(w, r, http.StatusBadRequest, "non absolute url")
		return
	}

	graph := app.crawlPage(urlStruct)
	env := envelope{"graph": graph}
	if err := app.writeJSON(w, env, http.StatusOK); err != nil {
		app.errResponse(w, r, http.StatusBadRequest, err.Error())
	}
}
