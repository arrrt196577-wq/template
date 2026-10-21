// Package app is the composition root. It owns shared dependencies and their cleanup.
package app

import (
	"context"
	"log/slog"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/httpserver"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/postgres"
)

type App struct {
	database *postgres.Connection
	server   *httpserver.Server
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	database, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}
	handler, err := httpserver.NewHandler(cfg.HTTP, log, database.SQL.PingContext)
	if err != nil {
		_ = database.Close()
		return nil, err
	}
	return &App{database: database, server: httpserver.New(cfg.HTTP, handler, log)}, nil
}

func (application *App) Run(ctx context.Context) error {
	return application.server.Run(ctx)
}

func (application *App) Close() error {
	return application.database.Close()
}
