package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/handlers"
)

func TestHealth_AlwaysOK(t *testing.T) {
	reg := handlers.NewHealthRegistry()
	rec := httptest.NewRecorder()
	reg.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want \"ok\"", body["status"])
	}
}

func TestReady_NoChecksRegistered(t *testing.T) {
	reg := handlers.NewHealthRegistry()
	rec := httptest.NewRecorder()
	reg.Ready(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 with no checks", rec.Code)
	}
}

func TestReady_CriticalDownReturns503(t *testing.T) {
	reg := handlers.NewHealthRegistry()
	reg.Register("postgres", true, func(context.Context) error { return errors.New("connection refused") })

	rec := httptest.NewRecorder()
	reg.Ready(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	deps := readyDeps(t, rec)
	if deps["postgres"] == "ok" {
		t.Errorf("postgres reported ok, want not ok")
	}
}

func TestReady_NonCriticalDegradedStaysReady(t *testing.T) {
	reg := handlers.NewHealthRegistry()
	reg.Register("postgres", true, func(context.Context) error { return nil })
	reg.Register("redis", false, func(context.Context) error { return errors.New("timeout") })

	rec := httptest.NewRecorder()
	reg.Ready(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 when only non-critical fails", rec.Code)
	}
	deps := readyDeps(t, rec)
	if deps["redis"] != "degraded" {
		t.Errorf("redis = %q, want \"degraded\"", deps["redis"])
	}
	if deps["postgres"] != "ok" {
		t.Errorf("postgres = %q, want \"ok\"", deps["postgres"])
	}
}

func readyDeps(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body struct {
		Status       string            `json:"status"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	return body.Dependencies
}
