package server

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("REQUEST BEGIN %s %s", r.Method, r.URL.Path)
		begin := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("REQUEST END %v", time.Since(begin))
	})
}
