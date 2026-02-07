package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				app.serverErrorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastseen time.Time
	}

	var (
		mutex   sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		time.Sleep(time.Minute)
		mutex.Lock()
		for ip, client := range clients {
			if time.Since(client.lastseen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mutex.Unlock()
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, http.StatusInternalServerError, err)
			}

			mutex.Lock()

			if c, ok := clients[ip]; ok {
				if !c.limiter.Allow() {
					mutex.Unlock()
					message := errors.New("rate limit exceeded")
					app.serverErrorResponse(w, r, http.StatusTooManyRequests, message)

					return
				}
			} else {
				clients[ip] = &client{
					limiter:  rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
					lastseen: time.Now(),
				}
			}

			mutex.Unlock()

		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) serverErrorResponse(
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
