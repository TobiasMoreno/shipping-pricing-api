// Command api is the entry point of the shipping-pricing-api service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TobiasMoreno/shipping-pricing-api/internal/config"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/handlers"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/server"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/services"
	"github.com/TobiasMoreno/shipping-pricing-api/internal/services/memory"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := newLogger(cfg.LogLevel)

	health := handlers.NewHealthRegistry()

	// In-memory repositories with seed data; replaced by PostgreSQL in a later change.
	zones, rules, promotions := memory.Seed()
	quoteService := services.NewQuoteService(rules, zones, promotions)
	shippingHandler := handlers.NewShippingHandler(quoteService)

	router := server.NewRouter(health, shippingHandler)
	srv := server.NewHTTPServer(fmt.Sprintf(":%d", cfg.HTTPPort), router)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "addr", srv.Addr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	logger.Info("server stopped cleanly")
	return nil
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
