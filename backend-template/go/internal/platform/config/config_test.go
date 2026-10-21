package config_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/config"
)

func load(values map[string]string) (config.Config, error) {
	return config.Load(func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	})
}

func valid() map[string]string {
	return map[string]string{"DB_NAME": "template_test", "DB_USERNAME": "test_user", "DB_PASSWORD": "test-secret"}
}

func TestConfigurationRejectsInvalidValuesWithoutEchoingSecrets(t *testing.T) {
	for _, test := range []struct{ key, value string }{
		{"DB_NAME", ""}, {"DB_USERNAME", ""}, {"DB_PASSWORD", ""},
		{"DB_PORT", "0"}, {"DB_MAX_OPEN_CONNS", "0"}, {"DB_MAX_IDLE_CONNS", "11"},
		{"HTTP_ADDR", "0.0.0.0:not-a-port"}, {"HTTP_READ_TIMEOUT", "1ms"},
		{"HTTP_SHUTDOWN_TIMEOUT", "0s"}, {"HTTP_WRITE_TIMEOUT", "1ms"},
		{"HTTP_CORS_ORIGINS", "*"}, {"HTTP_CORS_ORIGINS", "https://example.com/path"},
		{"HTTP_TRUSTED_PROXIES", "all"}, {"LOG_LEVEL", "bad-secret-value"},
		{"DB_SSLMODE", "bad-secret-value"}, {"APP_ENV", "bad-secret-value"},
	} {
		t.Run(test.key+"_"+test.value, func(t *testing.T) {
			values := valid()
			values[test.key] = test.value
			_, err := load(values)
			if err == nil || !strings.Contains(err.Error(), test.key) {
				t.Fatalf("expected an error naming %s", test.key)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("configuration error exposed a value")
			}
		})
	}
}

func TestDefaultsAndProductionTLS(t *testing.T) {
	cfg, err := load(valid())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.Address != "127.0.0.1:8081" || len(cfg.HTTP.TrustedProxies) != 0 || cfg.Database.SSLMode != "verify-full" {
		t.Fatal("unexpected defaults")
	}
	values := valid()
	values["APP_ENV"] = "production"
	values["DB_SSLMODE"] = "disable"
	if _, err := load(values); err == nil {
		t.Fatal("production accepted unverified plaintext transport")
	}
}

func TestDSNPreservesCredentialsAndDatabaseName(t *testing.T) {
	values := valid()
	values["DB_PASSWORD"] = "space @:/?#% secret"
	values["DB_USERNAME"] = "user@name"
	values["DB_NAME"] = "name with spaces"
	values["DB_HOST"] = "::1"
	cfg, err := load(values)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(cfg.Database.DSN())
	if err != nil {
		t.Fatal("invalid generated DSN")
	}
	password, _ := u.User.Password()
	if password != values["DB_PASSWORD"] || u.User.Username() != values["DB_USERNAME"] || u.Path != "/"+values["DB_NAME"] || u.Host != "[::1]:5432" {
		t.Fatal("DSN failed to preserve connection values")
	}
}
