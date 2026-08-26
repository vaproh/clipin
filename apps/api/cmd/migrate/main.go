package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"clipin/apps/api/internal/config"
)

func main() {
	flag.Parse()

	action := "up"
	if flag.NArg() > 0 {
		action = flag.Arg(0)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	m, err := migrator(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer m.Close()

	switch action {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "status":
		version, dirty, verr := m.Version()
		if verr != nil {
			log.Fatalf("Failed to read version: %v", verr)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
		return
	case "goto":
		if flag.NArg() < 2 {
			log.Fatal("Usage: migrate goto <version>")
		}
		v, perr := strconv.ParseUint(flag.Arg(1), 10, 32)
		if perr != nil {
			log.Fatalf("Invalid version %q: %v", flag.Arg(1), perr)
		}
		err = m.Migrate(uint(v))
	default:
		log.Fatalf("Unknown action: %s (use up, down, status, goto)", action)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Printf("Migration %s completed successfully\n", action)
}

// migrator builds a golang-migrate instance against db/migrations.
// The migrate pgx driver registers under the "pgx" scheme, so the connection
// URL is rewritten from the canonical postgres:// form.
func migrator(databaseURL string) (*migrate.Migrate, error) {
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
