package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Heavenly Demon God Domain</h1>")
}
