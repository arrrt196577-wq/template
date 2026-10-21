package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/app"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/buildinfo"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/logging"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		if err := json.NewEncoder(os.Stdout).Encode(buildinfo.Current()); err != nil {
			return 1
		}
		return 0
	}
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: server [--version]")
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
	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("application initialization failed", "reason", err.Error())
		return 1
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Error("database shutdown failed")
		}
	}()
	if err := application.Run(ctx); err != nil {
		log.Error("application stopped", "reason", err.Error())
		return 1
	}
	return 0
}
