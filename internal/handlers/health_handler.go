// Package handlers contains the HTTP handlers that translate requests into
// service calls and consistent JSON responses.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Check reports the health of a single dependency. A nil error means healthy.
type Check func(ctx context.Context) error

type checkEntry struct {
	name     string
	critical bool
	check    Check
}

// HealthRegistry holds the liveness and readiness handlers together with the
// set of dependency checks evaluated by readiness. Later changes register their
// dependencies (PostgreSQL, Redis, provider) without modifying the handlers.
type HealthRegistry struct {
	checks       []checkEntry
	checkTimeout time.Duration
}

// NewHealthRegistry creates an empty registry with a default per-check timeout.
func NewHealthRegistry() *HealthRegistry {
	return &HealthRegistry{checkTimeout: 2 * time.Second}
}

// Register adds a dependency check. Critical checks make readiness fail with
// 503; non-critical checks only mark the dependency as degraded.
func (r *HealthRegistry) Register(name string, critical bool, check Check) {
	r.checks = append(r.checks, checkEntry{name: name, critical: critical, check: check})
}

// Health is the liveness handler. It always returns 200 and never touches
// dependencies, so it reflects only whether the process is running.
func (r *HealthRegistry) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready is the readiness handler. It evaluates all registered checks and
// returns 503 when any critical dependency fails, 200 otherwise (with
// non-critical failures reported as "degraded").
func (r *HealthRegistry) Ready(w http.ResponseWriter, req *http.Request) {
	deps := make(map[string]string, len(r.checks))
	status := http.StatusOK

	for _, c := range r.checks {
		ctx, cancel := context.WithTimeout(req.Context(), r.checkTimeout)
		err := c.check(ctx)
		cancel()

		switch {
		case err == nil:
			deps[c.name] = "ok"
		case c.critical:
			deps[c.name] = "down"
			status = http.StatusServiceUnavailable
		default:
			deps[c.name] = "degraded"
		}
	}

	readyState := "ready"
	if status != http.StatusOK {
		readyState = "not_ready"
	}

	writeJSON(w, status, map[string]any{
		"status":       readyState,
		"dependencies": deps,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
