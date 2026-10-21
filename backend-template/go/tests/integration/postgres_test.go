//go:build integration

package integration_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/httpserver"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/migration"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/transaction"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Each test owns a randomly named schema and drops only that schema on cleanup.
func database(t *testing.T) (*gorm.DB, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_DSN is required for integration tests; use a dedicated test database")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("TEST_DATABASE_DSN must be a PostgreSQL URL")
	}
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal("cannot initialize test database")
	}
	t.Cleanup(func() { _ = admin.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := admin.PingContext(ctx); err != nil {
		t.Fatal("test database unavailable")
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	schema := "template_it_" + hex.EncodeToString(random[:])
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal("cannot create isolated test schema")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error("could not clean up isolated test schema")
		}
	})
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot connect to isolated test schema")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	return db, pool
}

func quiet() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestVersionedMigrationAndBusinessTransactionBoundary(t *testing.T) {
	db, pool := database(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	files := fstest.MapFS{
		"00001_fixture.sql": {Data: []byte("-- +goose Up\nCREATE TABLE transaction_probe (id BIGINT PRIMARY KEY, label TEXT NOT NULL);\n-- +goose Down\nDROP TABLE transaction_probe;\n")},
	}
	for range 2 {
		if err := migration.Run(ctx, pool, files, "up", quiet()); err != nil {
			t.Fatal(err)
		}
	}
	var applied int
	if err := pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM goose_db_version WHERE version_id = 1 AND is_applied").Scan(&applied); err != nil || applied != 1 {
		t.Fatal("repeated migration did not preserve a single applied version")
	}
	failure := errors.New("second operation failed")
	runner := transaction.New(db)
	err := runner.Within(ctx, func(tx *gorm.DB) error {
		if err := tx.Exec("INSERT INTO transaction_probe (id, label) VALUES (?, ?)", 1, "first").Error; err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatal("transaction did not return the operation failure")
	}
	var count int
	if err := pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM transaction_probe").Scan(&count); err != nil || count != 0 {
		t.Fatal("a failed transaction left a partial write")
	}
	if err := runner.Within(ctx, func(tx *gorm.DB) error {
		return tx.Exec("INSERT INTO transaction_probe (id, label) VALUES (?, ?), (?, ?)", 1, "first", 2, "second").Error
	}); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM transaction_probe").Scan(&count); err != nil || count != 2 {
		t.Fatal("a successful transaction did not commit")
	}
	files["00002_failed_fixture.sql"] = &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE failed_probe (id INT);\nSELECT * FROM nonexistent_fixture_table;\n")}
	if err := migration.Run(ctx, pool, files, "up", quiet()); err == nil {
		t.Fatal("invalid SQL migration was accepted")
	}
	var exists bool
	if err := pool.QueryRowContext(ctx, "SELECT to_regclass('failed_probe') IS NOT NULL").Scan(&exists); err != nil || exists {
		t.Fatal("failed SQL migration was not rolled back")
	}
}

func TestEmptyMigrationAndDatabaseReadiness(t *testing.T) {
	_, pool := database(t)
	ctx := context.Background()
	if err := migration.Run(ctx, pool, fstest.MapFS{"README.txt": {Data: []byte("no migrations")}}, "up", quiet()); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema()").Scan(&count); err != nil || count != 0 {
		t.Fatal("empty migration unexpectedly changed the database schema")
	}
	handler, err := httpserver.NewHandler(config.HTTP{HealthTimeout: time.Second}, quiet(), pool.PingContext)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/health/ready", nil))
	if response.Code != 200 {
		t.Fatal("connected database was not ready")
	}
	_ = pool.Close()
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/health/ready", nil))
	if response.Code != 503 {
		t.Fatal("closed database remained ready")
	}
}
