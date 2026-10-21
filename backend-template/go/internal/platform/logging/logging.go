// Package logging configures structured output without installing a global logger.
package logging

import (
	"io"
	"log/slog"
)

func New(output io.Writer, level slog.Level, service, environment string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, attribute slog.Attr) slog.Attr {
			if len(groups) == 0 && attribute.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, attribute.Value.Time().UTC().Format("2006-01-02T15:04:05.000Z"))
			}
			return attribute
		},
	})).With(
		"service", service,
		"environment", environment,
	)
}
