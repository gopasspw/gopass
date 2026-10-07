// Package buildinfo exposes the gopass build version to internal packages.
//
// The main package stashes the ldflags-injected version string here at
// startup; an empty value means unknown (unit tests that bypass main,
// library embedders). runtime/debug build info cannot serve this purpose:
// its Main.Version is "(devel)" for release builds, so
// debug.ModuleVersion("github.com/gopasspw/gopass") reliably resolves only
// dependency versions, not our own.
package buildinfo

import (
	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Version is the gopass build version, e.g. "1.17.3". Set by main at
// startup; empty when unset (test builds, library embedders).
var Version string

// ModuleVersion reports this binary's gopass version. The Version stash
// (set by main from the ldflags-injected version) is authoritative; the
// build-info module version is a fallback for embedders that bypass main
// and is unavailable in non-release builds. The second return value is
// false when no version could be determined. Unlike debug.ModuleVersion,
// which resolves a dependency's version by module path, this always refers
// to gopass itself.
func ModuleVersion() (string, bool) {
	if v := Version; v != "" {
		return v, true
	}
	if sv := debug.ModuleVersion("github.com/gopasspw/gopass"); !sv.Equals(semver.Version{}) {
		return sv.String(), true
	}

	return "", false
}
