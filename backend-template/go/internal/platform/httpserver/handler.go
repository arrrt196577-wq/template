// Package httpserver owns transport lifecycle, operational routes and middleware.
package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/gin-gonic/gin"
)

type Check func(context.Context) error

// NewHandler exposes only health endpoints. Business routing and authentication do not exist yet.
func NewHandler(cfg config.HTTP, log *slog.Logger, check Check) (http.Handler, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	// nil disables Gin's default trust of all proxies.
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}
	router.Use(requestID(), accessLog(log), recovery(log), cors(cfg.CORSOrigins))
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.HealthTimeout)
		defer cancel()
		if err := check(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})
	// Fail closed while no authentication or business routes have been implemented.
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "denied"})
	})
	return router, nil
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var id [16]byte
		if _, err := rand.Read(id[:]); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// Generate server-owned IDs, do not trust arbitrary incoming values.
		c.Header("X-Request-Id", hex.EncodeToString(id[:]))
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

func accessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		// Never record raw URL, query, body, headers or credentials.
		log.InfoContext(c.Request.Context(), "http request",
			"requestId", c.Writer.Header().Get("X-Request-Id"),
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"durationMs", time.Since(start).Milliseconds(),
		)
	}
}

func recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recover() != nil {
				// Panic values and request dumps can contain secrets.
				log.ErrorContext(c.Request.Context(), "http handler panic",
					"requestId", c.Writer.Header().Get("X-Request-Id"))
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error"})
				} else {
					c.Abort()
				}
			}
		}()
		c.Next()
	}
}

func cors(origins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		allowed[origin] = true
	}
	methods := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "OPTIONS": true}
	headers := map[string]bool{"authorization": true, "content-type": true}
	return func(c *gin.Context) {
		// Vary also on responses without Origin, so caches cannot mix CORS variants.
		c.Writer.Header().Add("Vary", "Origin")
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		if !allowed[origin] {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Expose-Headers", "X-Request-Id")
		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			c.Writer.Header().Add("Vary", "Access-Control-Request-Method")
			c.Writer.Header().Add("Vary", "Access-Control-Request-Headers")
			if !methods[c.GetHeader("Access-Control-Request-Method")] {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			for name := range strings.SplitSeq(c.GetHeader("Access-Control-Request-Headers"), ",") {
				if name = strings.ToLower(strings.TrimSpace(name)); name != "" && !headers[name] {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Max-Age", "600")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
