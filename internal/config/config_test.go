package config_test

import (
	"os"
	"testing"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/config"
)

func TestLoad_AppliesDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/shipping")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("CACHE_ENABLED")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %d, want 8080 (default)", cfg.HTTPPort)
	}
	if !cfg.CacheEnabled {
		t.Error("CacheEnabled = false, want true (default)")
	}
}

func TestLoad_ReadsFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/shipping")
	t.Setenv("HTTP_PORT", "9000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != 9000 {
		t.Errorf("HTTPPort = %d, want 9000", cfg.HTTPPort)
	}
}

func TestLoad_MissingRequiredFails(t *testing.T) {
	os.Unsetenv("DATABASE_URL")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing, got nil")
	}
}

func TestLoad_InvalidPortFails(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/shipping")
	t.Setenv("HTTP_PORT", "not-a-number")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when HTTP_PORT is non-numeric, got nil")
	}
}
