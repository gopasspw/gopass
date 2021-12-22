package debug

import (
	rdebug "runtime/debug"
	"strings"

	"github.com/blang/semver/v4"
)

// ModuleVersion the version of the named import
func ModuleVersion(m string) semver.Version {
	bi, ok := rdebug.ReadBuildInfo()
	if !ok || bi == nil {
		Log("Failed to read build info")
		return semver.Version{}
	}

	for _, dep := range bi.Deps {
		if dep.Path != m {
			continue
		}
		sv, err := semver.Parse(strings.TrimPrefix(dep.Version, "v"))
		if err != nil {
			Log("Failed to parse version %s: %s", dep.Version, err)
			return semver.Version{}
		}
		return sv
	}
	Log("no module %s found", m)
	return semver.Version{}
}
