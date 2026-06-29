// Package migrate runs the embedded database migrations using golang-migrate as
// a library (no external CLI required).
package migrate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // register pgx5:// driver
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/TobiasMoreno/shipping-pricing-api/migrations"
)

// Up applies all pending migrations.
func Up(databaseURL string) error { return run(databaseURL, true) }

// Down rolls back all migrations.
func Down(databaseURL string) error { return run(databaseURL, false) }

func run(databaseURL string, up bool) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migrate: load source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, toPgxURL(databaseURL))
	if err != nil {
		return fmt.Errorf("migrate: init: %w", err)
	}
	defer m.Close()

	if up {
		err = m.Up()
	} else {
		err = m.Down()
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: apply: %w", err)
	}
	return nil
}

// toPgxURL rewrites a postgres:// DSN to the pgx5:// scheme expected by the
// golang-migrate pgx/v5 driver.
func toPgxURL(u string) string {
	switch {
	case strings.HasPrefix(u, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(u, "postgres://")
	case strings.HasPrefix(u, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(u, "postgresql://")
	default:
		return u
	}
}
