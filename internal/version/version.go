package version

import (
	"runtime"
	"runtime/debug"
)

// These variables are set at build time using -ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	Date      = "unknown"
	GoVersion = runtime.Version()
	Platform  = runtime.GOOS + "/" + runtime.GOARCH
)

// init applies a best-effort fallback for version metadata when ldflags are not set.
// This improves the output for binaries installed via `go install module@version`.
func init() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		// If Version wasn't injected, prefer the module's semantic version when available.
		if Version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			Version = bi.Main.Version
		}
		// Try to populate Commit and Date from VCS settings if present (best effort).
		// Note: these settings are typically present when building from a VCS checkout.
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if Commit == "none" && s.Value != "" {
					if len(s.Value) >= 7 {
						Commit = s.Value[:7]
					} else {
						Commit = s.Value
					}
				}
			case "vcs.time":
				if Date == "unknown" && s.Value != "" {
					Date = s.Value
				}
			}
		}
	}
}
