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

	r.NotFound(app.notFoundResponse)
	r.MethodNotAllowed(app.methodNotAllowedResponse)

	r.Get("/healthcheck", app.healthcheckHandler)
	r.Post("/graph", app.getGraph)

	return r
}
