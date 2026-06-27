// Package server wires the HTTP router and server lifecycle.
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/handlers"
)

// NewRouter builds the application router, wiring middlewares and routes.
func NewRouter(health *handlers.HealthRegistry, shipping *handlers.ShippingHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(requestIDResponseHeader)
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Health)
	r.Get("/ready", health.Ready)

	r.Post("/shipping/quote", shipping.Quote)

	return r
}

// requestIDResponseHeader echoes the request id (set by middleware.RequestID)
// back to the client in the X-Request-ID response header.
func requestIDResponseHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := middleware.GetReqID(r.Context()); id != "" {
			w.Header().Set("X-Request-ID", id)
		}
		next.ServeHTTP(w, r)
	})
}
