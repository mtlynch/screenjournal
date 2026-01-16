package sqlite

import (
	"context"
	"embed"
	"io/fs"
	"log"

	migrate "codeberg.org/mtlynch/go-evolutionary-migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func (s Store) applyMigrations() {
	migrationsRoot, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("failed to load migration files: %v", err)
	}

	if err := migrate.Run(context.Background(), s.ctx, migrationsRoot); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}
}
