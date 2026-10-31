package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
)

func TestApplicationRoutesExposeOnlyHealth(t *testing.T) {
	var logs bytes.Buffer
	var checks int
	var unhealthy bool
	handler, err := newHTTPHandler(config.HTTP{HealthTimeout: time.Second}, slog.New(slog.NewJSONHandler(&logs, nil)), func(context.Context) error {
		checks++
		if unhealthy {
			return errors.New("database-secret")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/health/live", http.StatusOK},
		{http.MethodGet, "/health/ready", http.StatusOK},
		{http.MethodPost, "/health/live", http.StatusForbidden},
		{http.MethodGet, "/api/v1/users", http.StatusForbidden},
		{http.MethodPost, "/api/v1/users", http.StatusForbidden},
		{http.MethodGet, "/api/v1/me", http.StatusForbidden},
		{http.MethodPost, "/api/v1/auth/login", http.StatusForbidden},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
		if response.Code != test.status {
			t.Fatalf("%s %s: got %d, want %d", test.method, test.path, response.Code, test.status)
		}
		if len(response.Header().Get("X-Request-Id")) != 32 || response.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("application route bypassed transport middleware")
		}
	}
	if checks != 1 {
		t.Fatalf("only readiness should query the database; got %d checks", checks)
	}
	unhealthy = true
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if response.Code != http.StatusServiceUnavailable || strings.Contains(response.Body.String()+logs.String(), "database-secret") {
		t.Fatal("readiness must report failure without exposing database diagnostics")
	}
}

func TestApplicationRoutesRejectInvalidProxyConfiguration(t *testing.T) {
	handler, err := newHTTPHandler(config.HTTP{TrustedProxies: []string{"invalid-proxy"}}, slog.Default(), func(context.Context) error { return nil })
	if err == nil || handler != nil {
		t.Fatal("application must not serve with invalid trusted proxies")
	}
}
