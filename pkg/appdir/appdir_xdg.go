//go:build !windows

package appdir

import (
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/debug"
)

// UserConfig returns the user's config directory.
// It follows the XDG Base Directory Specification.
// The GOPASS_HOMEDIR environment variable can be used to override the base path.
// See: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func (a *Appdir) UserConfig() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		debug.V(3).Log("GOPASS_HOMEDIR is set to %s", hd)

		return filepath.Join(hd, ".config", a.name)
	}

	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}

	return filepath.Join(base, a.name)
}

// UserCache returns the user's cache directory.
// It follows the XDG Base Directory Specification.
// The GOPASS_HOMEDIR environment variable can be used to override the base path.
// See: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func (a *Appdir) UserCache() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".cache", a.name)
	}

	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".cache")
	}

	return filepath.Join(base, a.name)
}

// UserData returns the user's data directory.
// It follows the XDG Base Directory Specification.
// The GOPASS_HOMEDIR environment variable can be used to override the base path.
// See: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func (a *Appdir) UserData() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".local", "share", a.name)
	}

	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}

	return filepath.Join(base, a.name)
}
