// Package migrations embeds the Go backend's SQL assets, independent of working directory.
package migrations

import (
	"embed"
	"io/fs"
)

//go:embed sql
var assets embed.FS

func Files() (fs.FS, error) {
	return fs.Sub(assets, "sql")
}
