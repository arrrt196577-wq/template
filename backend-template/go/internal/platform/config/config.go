// Package config reads and validates process configuration without global state.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	ServiceName string
	LogLevel    slog.Level
	HTTP        HTTP
	Database    Database
}

type HTTP struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	HealthTimeout     time.Duration
	CORSOrigins       []string
	TrustedProxies    []string
}

type Database struct {
	Host            string
	Port            int
	Name            string
	Username        string
	Password        string
	SSLMode         string
	SSLRootCert     string
	ConnectTimeout  time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Load accepts a lookup function so tests and commands need not mutate global environment.
// Errors contain variable names, never their supplied values.
func Load(lookup func(string) (string, bool)) (Config, error) {
	var cfg Config
	var firstErr error
	value := func(key, fallback string) string {
		v, ok := lookup(key)
		if !ok {
			return fallback
		}
		return v
	}
	fail := func(key string) {
		if firstErr == nil {
			firstErr = fmt.Errorf("invalid or missing configuration: %s", key)
		}
	}
	required := func(key string) string {
		v := value(key, "")
		if strings.TrimSpace(v) == "" {
			fail(key)
		}
		return v
	}
	duration := func(key, fallback string) time.Duration {
		v, err := time.ParseDuration(value(key, fallback))
		if err != nil || v <= 0 {
			fail(key)
		}
		return v
	}
	integer := func(key, fallback string, min, max int) int {
		v, err := strconv.Atoi(value(key, fallback))
		if err != nil || v < min || v > max {
			fail(key)
		}
		return v
	}
	cfg.Environment = value("APP_ENV", "development")
	switch cfg.Environment {
	case "development", "test", "production":
	default:
		fail("APP_ENV")
	}
	cfg.ServiceName = value("SERVICE_NAME", "template-go")
	if cfg.ServiceName == "" || strings.ContainsAny(cfg.ServiceName, "\r\n") {
		fail("SERVICE_NAME")
	}
	if err := cfg.LogLevel.UnmarshalText([]byte(value("LOG_LEVEL", "INFO"))); err != nil {
		fail("LOG_LEVEL")
	}
	cfg.HTTP.Address = value("HTTP_ADDR", "127.0.0.1:8081")
	host, port, err := net.SplitHostPort(cfg.HTTP.Address)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || portNumber < 1 || portNumber > 65535 || (host != "" && net.ParseIP(host) == nil && host != "localhost") {
		fail("HTTP_ADDR")
	}
	cfg.HTTP.ReadHeaderTimeout = duration("HTTP_READ_HEADER_TIMEOUT", "5s")
	cfg.HTTP.ReadTimeout = duration("HTTP_READ_TIMEOUT", "15s")
	cfg.HTTP.WriteTimeout = duration("HTTP_WRITE_TIMEOUT", "30s")
	cfg.HTTP.IdleTimeout = duration("HTTP_IDLE_TIMEOUT", "60s")
	cfg.HTTP.ShutdownTimeout = duration("HTTP_SHUTDOWN_TIMEOUT", "10s")
	cfg.HTTP.HealthTimeout = duration("HTTP_HEALTH_TIMEOUT", "2s")
	if cfg.HTTP.ReadTimeout < cfg.HTTP.ReadHeaderTimeout {
		fail("HTTP_READ_TIMEOUT")
	}
	if cfg.HTTP.WriteTimeout <= cfg.HTTP.HealthTimeout {
		fail("HTTP_WRITE_TIMEOUT")
	}
	cfg.HTTP.CORSOrigins = list(value("HTTP_CORS_ORIGINS", ""))
	for _, origin := range cfg.HTTP.CORSOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			fail("HTTP_CORS_ORIGINS")
		}
	}
	cfg.HTTP.TrustedProxies = list(value("HTTP_TRUSTED_PROXIES", ""))
	for _, proxy := range cfg.HTTP.TrustedProxies {
		_, _, err := net.ParseCIDR(proxy)
		if err != nil && net.ParseIP(proxy) == nil {
			fail("HTTP_TRUSTED_PROXIES")
		}
	}
	cfg.Database = Database{
		Host:            value("DB_HOST", "127.0.0.1"),
		Port:            integer("DB_PORT", "5432", 1, 65535),
		Name:            required("DB_NAME"),
		Username:        required("DB_USERNAME"),
		Password:        required("DB_PASSWORD"),
		SSLMode:         value("DB_SSLMODE", "verify-full"),
		SSLRootCert:     value("DB_SSLROOTCERT", ""),
		ConnectTimeout:  duration("DB_CONNECT_TIMEOUT", "5s"),
		MaxOpenConns:    integer("DB_MAX_OPEN_CONNS", "10", 1, 1000),
		MaxIdleConns:    integer("DB_MAX_IDLE_CONNS", "5", 0, 1000),
		ConnMaxLifetime: duration("DB_CONN_MAX_LIFETIME", "30m"),
		ConnMaxIdleTime: duration("DB_CONN_MAX_IDLE_TIME", "5m"),
	}
	if strings.TrimSpace(cfg.Database.Host) == "" {
		fail("DB_HOST")
	}
	if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		fail("DB_MAX_IDLE_CONNS")
	}
	switch cfg.Database.SSLMode {
	case "disable", "require", "verify-ca", "verify-full":
	default:
		fail("DB_SSLMODE")
	}
	if cfg.Environment == "production" && cfg.Database.SSLMode != "verify-full" {
		fail("DB_SSLMODE")
	}
	return cfg, firstErr
}

func list(value string) []string {
	var result []string
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

// DSN must never be logged. URL encoding preserves special characters in credentials.
func (cfg Database) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:   "/" + cfg.Name,
	}
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)
	query.Set("connect_timeout", strconv.FormatInt(max(1, int64(cfg.ConnectTimeout.Round(time.Second)/time.Second)), 10))
	if cfg.SSLRootCert != "" {
		query.Set("sslrootcert", cfg.SSLRootCert)
	}
	u.RawQuery = query.Encode()
	return u.String()
}
