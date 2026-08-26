package db

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateUp applies all pending migrations to the given database.
func MigrateUp(databaseURL string) error {
	m, err := openMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// openMigrator builds a golang-migrate instance against db/migrations.
// The migrate pgx driver registers under the "pgx" scheme, so the connection
// URL is rewritten from the canonical postgres:// form.
func openMigrator(databaseURL string) (*migrate.Migrate, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	u.Scheme = "pgx"

	m, err := migrate.New("file://db/migrations", u.String())
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}
	return m, nil
}
