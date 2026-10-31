// Package httpserver owns transport lifecycle, operational routes and middleware.
package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/gin-gonic/gin"
)

// NewRouter configures transport policies without registering application routes.
// Routes must be registered before serving requests; authentication is not implemented.
func NewRouter(cfg config.HTTP, log *slog.Logger) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// nil disables Gin's default trust of all proxies.
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}
	router.Use(requestID(), accessLog(log), recovery(log), cors(cfg.CORSOrigins))
	// Unknown routes remain denied until the common API error contract is implemented.
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "denied"})
	})
	return router, nil
}
