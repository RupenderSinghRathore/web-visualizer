package main

import "net/http"

func (app *application) errResponse(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	err error,
) {
	http.Error(w, "", status)
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err, "method", method, "uri", uri)
}
