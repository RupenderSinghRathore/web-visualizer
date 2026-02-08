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

	r.Get("/healthcheck", app.healthcheckHandler)
	r.Get("/", app.homePageHandler)

	return r
}
