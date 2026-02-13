package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(app.recoverPanic)
	r.Use(app.rateLimit)

	fs := http.FileServer(http.Dir("./web/assets/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/healthcheck", app.healthcheckHandler)

	// r.Get("/", app.homePageHandler)
	// r.Post("/graph", app.drawGraphHandler)

	return r
}
