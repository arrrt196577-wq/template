package app

import (
	"log/slog"
	"net/http"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/httpserver"
)

// newHTTPHandler composes application routes with the shared transport policies.
// Business HTTP adapters will be wired here once their contracts and services exist.
func newHTTPHandler(cfg config.HTTP, log *slog.Logger, check httpserver.Check) (http.Handler, error) {
	router, err := httpserver.NewRouter(cfg, log)
	if err != nil {
		return nil, err
	}
	httpserver.RegisterHealthRoutes(router, cfg.HealthTimeout, check)
	return router, nil
}
