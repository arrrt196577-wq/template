// Package buildinfo reports the toolchain and source revision recorded by go build.
package buildinfo

import (
	"runtime"
	"runtime/debug"
)

type Info struct {
	GoVersion string `json:"goVersion"`
	Revision  string `json:"revision"`
	Modified  bool   `json:"modified"`
}

func Current() Info {
	info := Info{GoVersion: runtime.Version(), Revision: "unknown"}
	if build, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range build.Settings {
			switch setting.Key {
			case "vcs.revision":
				info.Revision = setting.Value
			case "vcs.modified":
				info.Modified = setting.Value == "true"
			}
		}
	}
	return info
}
