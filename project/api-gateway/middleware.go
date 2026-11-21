package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Middleware для логирования HTTP-запросов.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf(
			"%s %s duration=%s",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}

// Middleware для ограничения времени обработки HTTP-запроса.
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
