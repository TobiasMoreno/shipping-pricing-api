// Package config loads and validates the application configuration from
// environment variables into a typed struct.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the typed application configuration.
type Config struct {
	AppEnv   string
	HTTPPort int
	LogLevel string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	CacheEnabled         bool
	QuoteCacheTTLSeconds int
	RulesCacheTTLSeconds int

	ProviderBaseURL    string
	ProviderTimeoutMS  int
	ProviderMaxRetries int

	RateLimitRequestsPerMinute int
}

// Load reads configuration from the environment, applying defaults for optional
// values and validating required ones. It returns an error (so the process can
// fail fast at startup) when a required value is missing or malformed.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:        getString("APP_ENV", "local"),
		LogLevel:      getString("LOG_LEVEL", "info"),
		DatabaseURL:   getString("DATABASE_URL", ""),
		RedisAddr:     getString("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getString("REDIS_PASSWORD", ""),
		CacheEnabled:  getBool("CACHE_ENABLED", true),
		ProviderBaseURL: getString("PROVIDER_BASE_URL", ""),
	}

	var err error
	if cfg.HTTPPort, err = getInt("HTTP_PORT", 8080); err != nil {
		return Config{}, err
	}
	if cfg.RedisDB, err = getInt("REDIS_DB", 0); err != nil {
		return Config{}, err
	}
	if cfg.QuoteCacheTTLSeconds, err = getInt("QUOTE_CACHE_TTL_SECONDS", 600); err != nil {
		return Config{}, err
	}
	if cfg.RulesCacheTTLSeconds, err = getInt("RULES_CACHE_TTL_SECONDS", 3600); err != nil {
		return Config{}, err
	}
	if cfg.ProviderTimeoutMS, err = getInt("PROVIDER_TIMEOUT_MS", 800); err != nil {
		return Config{}, err
	}
	if cfg.ProviderMaxRetries, err = getInt("PROVIDER_MAX_RETRIES", 2); err != nil {
		return Config{}, err
	}
	if cfg.RateLimitRequestsPerMinute, err = getInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 120); err != nil {
		return Config{}, err
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// validate enforces required configuration invariants.
func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("config: HTTP_PORT must be in range 1-65535, got %d", c.HTTPPort)
	}
	return nil
}

func getString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getInt(key string, def int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be an integer, got %q", key, v)
	}
	return n, nil
}

func getBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
