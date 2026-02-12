package main

import (
	"RupenderSinghRathore/web-visualizer/web/view"
	"RupenderSinghRathore/web-visualizer/web/view/pages"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Heavenly Demon God Domain</h1>")
}

func (app *application) homePageHandler(w http.ResponseWriter, r *http.Request) {
	view.Base(pages.Home()).Render(context.Background(), w)
}

func (app *application) drawGraphHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.errResponse(w, r, http.StatusBadRequest, err)
	}
	endpoint := r.PostForm.Get("url")
	validUrl, err := url.ParseRequestURI(endpoint)
	if err != nil {
		app.errResponse(w, r, http.StatusBadRequest, err)
	}
	graph, err := app.crawlPage(validUrl)

	if err != nil {
	}
	pages.DrawMapTree(graph, validUrl.Path).Render(context.Background(), w)
	pages.DrawGraph(graph).Render(context.Background(), w)
}
