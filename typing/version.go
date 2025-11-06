package typing

import (
	"fmt"
	"strconv"
	"strings"
)

// ContractVersion is the version of the plugin contract the host is built against.
// Bump MAJOR for breaking changes, MINOR for backwards-compatible additions, PATCH for fixes.
const ContractVersion = "2.0.0"

// MinSupportedContractVersion expresses the minimum contract version the host will accept.
// Update this when dropping support for older contract versions.
const MinSupportedContractVersion = "2.0.0"

// ParseSemVer turns a semver-like string into (major, minor, patch), ignoring pre-release/build.
func ParseSemVer(v string) (int, int, int) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, 0, 0
	}
	// Remove pre-release / build metadata
	if i := strings.IndexAny(v, "+-"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	get := func(i int) int {
		if i >= len(parts) {
			return 0
		}
		n, _ := strconv.Atoi(parts[i])
		return n
	}
	return get(0), get(1), get(2)
}

// CompareSemVer compares two semver-like strings. Returns -1 if a<b, 0 if equal, 1 if a>b.
func CompareSemVer(a, b string) int {
	am, an, ap := ParseSemVer(a)
	bm, bn, bp := ParseSemVer(b)
	if am != bm {
		if am < bm {
			return -1
		}
		return 1
	}
	if an != bn {
		if an < bn {
			return -1
		}
		return 1
	}
	if ap != bp {
		if ap < bp {
			return -1
		}
		return 1
	}
	return 0
}

// IsCompatible checks whether a plugin contract version is compatible with the host.
// Basic policy:
//   - plugin MAJOR must equal host ContractVersion MAJOR
//   - plugin version must be >= MinSupportedContractVersion
func IsCompatible(pluginVersion string) bool {
	pm, _, _ := ParseSemVer(pluginVersion)
	hm, _, _ := ParseSemVer(ContractVersion)
	if pm != hm {
		return false
	}
	return CompareSemVer(pluginVersion, MinSupportedContractVersion) >= 0
}

// IncompatibilityMessage returns a concise reason for incompatibility.
func IncompatibilityMessage(pluginVersion string) string {
	return fmt.Sprintf("host contract %s, plugin contract %s (min supported %s)", ContractVersion, pluginVersion, MinSupportedContractVersion)
}
