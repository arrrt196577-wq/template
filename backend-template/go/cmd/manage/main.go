package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/buildinfo"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/logging"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/migration"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/postgres"
	"github.com/arrrt196577-wq/template/backend-template/go/migrations"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 1 && args[0] == "--version" {
		if err := json.NewEncoder(os.Stdout).Encode(buildinfo.Current()); err != nil {
			return 1
		}
		return 0
	}
	if len(args) != 2 || args[0] != "migrate" || (args[1] != "up" && args[1] != "status") {
		fmt.Fprintln(os.Stderr, "usage: manage migrate <up|status>; administrator initialization is not implemented")
		return 2
	}
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	log := logging.New(os.Stdout, cfg.LogLevel, cfg.ServiceName, cfg.Environment)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	database, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		log.Error("management command initialization failed", "reason", err.Error())
		return 1
	}
	defer func() { _ = database.Close() }()
	files, err := migrations.Files()
	if err != nil {
		log.Error("embedded migration source unavailable")
		return 1
	}
	if err := migration.Run(ctx, database.SQL, files, args[1], log); err != nil {
		log.Error("management command failed", "reason", err.Error())
		return 1
	}
	return 0
}
