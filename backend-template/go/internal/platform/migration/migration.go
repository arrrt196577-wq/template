// Package migration runs explicit, locked SQL migrations without global goose state.
package migration

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"log/slog"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// Run supports only forward migration and status inspection. Destructive commands are not exposed.
// Error details from SQL are intentionally not logged by this layer.
func Run(ctx context.Context, db *sql.DB, files fs.FS, action string, log *slog.Logger) error {
	if action != "up" && action != "status" {
		return errors.New("unsupported migration command; use up or status")
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return errors.New("migration lock initialization failed")
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, files,
		goose.WithDisableGlobalRegistry(true),
		goose.WithSessionLocker(locker),
	)
	if errors.Is(err, goose.ErrNoMigrations) {
		log.Info("no SQL migrations embedded; database schema unchanged")
		return nil
	}
	if err != nil {
		return errors.New("migration source validation failed")
	}
	switch action {
	case "up":
		results, err := provider.Up(ctx)
		if err != nil {
			return errors.New("database migration failed; inspect database diagnostics")
		}
		log.Info("database migrations completed", "applied", len(results))
	case "status":
		statuses, err := provider.Status(ctx)
		if err != nil {
			return errors.New("database migration status unavailable")
		}
		for _, status := range statuses {
			log.Info("database migration status", "version", status.Source.Version, "state", status.State)
		}
	}
	return nil
}
