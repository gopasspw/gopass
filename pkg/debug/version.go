package debug

import (
	rdebug "runtime/debug"
	"strings"

	"github.com/blang/semver/v4"
)

var biFunc func() (*rdebug.BuildInfo, bool) = rdebug.ReadBuildInfo

// ModuleVersion the version of the named import.
func ModuleVersion(m string) semver.Version {
	bi, ok := biFunc()
	if !ok || bi == nil {
		Log("Failed to read build info")

		return semver.Version{}
	}

	// special case for gopass
	if m == "github.com/gopasspw/gopass" || strings.HasPrefix(m, "github.com/gopasspw/gopass/") {
		sv, err := semver.Parse(strings.TrimPrefix(bi.Main.Version, "v"))
		if err == nil {
			return sv
		}
		Log("Failed to parse version %q for %q (gopass): %s", bi.Main.Version, m, err)
	}

	for _, dep := range bi.Deps {
		// We might be asking for a package that is part of a module
		// but not the module itself.
		if !strings.HasPrefix(m, dep.Path) {
			continue
		}

		sv, err := semver.Parse(strings.TrimPrefix(dep.Version, "v"))
		if err != nil {
			Log("Failed to parse version %q for %q: %s", dep.Version, dep.Path, err)

			if dep.Version == "" {
				return semver.Version{}
			}

			// remove invalid characters
			dv := strings.Trim(strings.TrimPrefix(dep.Version, "v"), "()")

			return semver.Version{
				Build: []string{dv},
			}
		}

		return sv
	}

	Log("no module %s found. Modules: %v", m, paths(bi.Deps))

	return semver.Version{}
}

func paths(mods []*rdebug.Module) []string {
	out := make([]string, 0, len(mods))
	for _, m := range mods {
		out = append(out, m.Path)
	}

	return out
}
