// Package postgres owns connection setup and pool configuration, not business queries.
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Connection struct {
	ORM *gorm.DB
	SQL *sql.DB
}

func Open(ctx context.Context, cfg config.Database) (*Connection, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		DisableAutomaticPing: true,
		TranslateError:       true,
		// Driver errors and SQL may contain credentials or bound values.
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, errors.New("PostgreSQL driver initialization failed")
	}
	pool, err := db.DB()
	if err != nil {
		return nil, errors.New("PostgreSQL connection pool initialization failed")
	}
	pool.SetMaxOpenConns(cfg.MaxOpenConns)
	pool.SetMaxIdleConns(cfg.MaxIdleConns)
	pool.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	pool.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := pool.PingContext(pingCtx); err != nil {
		_ = pool.Close()
		return nil, errors.New("PostgreSQL connection check failed; verify configuration and availability")
	}
	return &Connection{ORM: db, SQL: pool}, nil
}

func (connection *Connection) Close() error {
	return connection.SQL.Close()
}
