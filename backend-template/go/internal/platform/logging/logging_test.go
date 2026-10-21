package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/arrrt196577-wq/template/backend-template/go/internal/platform/logging"
)

func TestStructuredLogUsesUTCMillisecondsAndServiceIdentity(t *testing.T) {
	var buffer bytes.Buffer
	log := logging.New(&buffer, slog.LevelInfo, "test-service", "test")
	log.Debug("not enabled")
	log.Info("started")
	var entry map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	stamp, ok := entry["time"].(string)
	if !ok || !strings.HasSuffix(stamp, "Z") || len(stamp) != 24 {
		t.Fatal("log timestamp is not UTC with fixed millisecond precision")
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000Z", stamp); err != nil {
		t.Fatal(err)
	}
	if entry["service"] != "test-service" || entry["environment"] != "test" || entry["msg"] != "started" {
		t.Fatal("missing log identity or incorrect severity filtering")
	}
}
