package httpserver

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
	"github.com/gin-gonic/gin"
)

func TestHealthFailClosedAndLogRedaction(t *testing.T) {
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	var unhealthy bool
	handler, err := NewHandler(config.HTTP{HealthTimeout: time.Second}, log, func(context.Context) error {
		if unhealthy {
			return errors.New("internal database-secret")
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
		{"GET", "/health/live", 200}, {"GET", "/health/ready", 200},
		{"POST", "/health/live", 403}, {"GET", "/api/v1/users?token=query-secret", 403},
		{"POST", "/auth/login", 403}, {"GET", "/path-secret", 403},
	} {
		request := httptest.NewRequest(test.method, test.path, strings.NewReader("body-secret"))
		request.Header.Set("Authorization", "Bearer header-secret")
		request.Header.Set("X-Request-Id", "untrusted-secret")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.status || len(response.Header().Get("X-Request-Id")) != 32 {
			t.Fatalf("%s %s: unexpected status or request ID", test.method, test.path)
		}
	}
	unhealthy = true
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/health/ready", nil))
	if response.Code != 503 || strings.Contains(response.Body.String(), "secret") {
		t.Fatal("readiness must fail without exposing connection diagnostics")
	}
	if strings.Contains(logs.String(), "secret") {
		t.Fatal("request logs exposed credentials or untrusted URL data")
	}
}

func TestReadinessHonorsDeadlineAndRecoveryDoesNotExposePanic(t *testing.T) {
	var logs bytes.Buffer
	handler, err := NewHandler(config.HTTP{HealthTimeout: time.Millisecond}, slog.New(slog.NewJSONHandler(&logs, nil)), func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/health/ready", nil))
	if response.Code != 503 {
		t.Fatal("readiness ignored its timeout")
	}
	handler.(*gin.Engine).GET("/panic-test", func(*gin.Context) { panic("panic-secret") })
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/panic-test", nil))
	if response.Code != 500 || strings.Contains(logs.String()+response.Body.String(), "panic-secret") {
		t.Fatal("panic was not recovered safely")
	}
}

func TestCORSAndProxyTrust(t *testing.T) {
	handler, err := NewHandler(config.HTTP{HealthTimeout: time.Second, CORSOrigins: []string{"https://admin.example.com"}}, slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), func(context.Context) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		origin, headers string
		status          int
	}{
		{"https://admin.example.com", "Authorization, Content-Type", 204},
		{"https://attacker.example.com", "Authorization", 403},
		{"https://admin.example.com", "X-Unapproved-Header", 403},
	} {
		request := httptest.NewRequest("OPTIONS", "/future-api", nil)
		request.Header.Set("Origin", test.origin)
		request.Header.Set("Access-Control-Request-Method", "POST")
		request.Header.Set("Access-Control-Request-Headers", test.headers)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.status || response.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Fatal("unexpected CORS policy")
		}
	}
	var clientIP string
	handler.(*gin.Engine).GET("/proxy-test", func(c *gin.Context) { clientIP = c.ClientIP(); c.Status(204) })
	request := httptest.NewRequest("GET", "/proxy-test", nil)
	request.RemoteAddr = "192.0.2.1:1234"
	request.Header.Set("X-Forwarded-For", "198.51.100.1")
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if clientIP != "192.0.2.1" {
		t.Fatal("trusted a forged forwarding header")
	}
}
