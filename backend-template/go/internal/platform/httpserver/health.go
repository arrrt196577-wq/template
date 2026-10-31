package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Check func(context.Context) error

// RegisterHealthRoutes registers public operational endpoints, not business APIs.
func RegisterHealthRoutes(routes gin.IRoutes, timeout time.Duration, check Check) {
	routes.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})
	routes.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		if err := check(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})
}
