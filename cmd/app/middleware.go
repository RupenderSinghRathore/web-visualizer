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
				app.serverErrResponse(w, r, fmt.Errorf("%v", err))
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
				app.serverErrResponse(w, r, err)
			}

			mutex.Lock()

			if c, ok := clients[ip]; ok {
				if !c.limiter.Allow() {
					mutex.Unlock()
					message := errors.New("rate limit exceeded")
					app.errResponse(w, r, http.StatusTooManyRequests, message.Error())

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

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().
			Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
