package main

import (
	"RupenderSinghRathore/web-visualizer/web/view"
	"RupenderSinghRathore/web-visualizer/web/view/pages"
	"context"
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Heavenly Demon God Domain</h1>")
}

func (app *application) homePageHandler(w http.ResponseWriter, r *http.Request) {
	view.Base(pages.Home(), nil).Render(context.Background(), w)
}
