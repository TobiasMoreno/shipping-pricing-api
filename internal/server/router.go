// Package server wires the HTTP router and server lifecycle.
package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/handlers"
)

// NewRouter builds the application router and mounts the health endpoints.
// Business routes and middlewares are added by later changes.
func NewRouter(health *handlers.HealthRegistry) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/health", health.Health)
	r.Get("/ready", health.Ready)

	return r
}
