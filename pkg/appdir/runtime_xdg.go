//go:build !windows
// +build !windows

package appdir

import (
	"os"
	"path/filepath"
)

// UserRuntime returns the users runtime dir.
func (a *Appdir) UserRuntime() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".run")
	}

	base := os.Getenv("XDG_RUNTIME_DIR")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".run")
	}

	return filepath.Join(base, a.name)
}

// UserRuntime returns the users runtime dir.
func UserRuntime() string {
	return DefaultAppdir.UserRuntime()
}
