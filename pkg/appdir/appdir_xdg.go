//go:build !windows
// +build !windows

package appdir

import (
	"os"
	"path/filepath"
)

// UserConfig returns the users config dir.
func (a *Appdir) UserConfig() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".config", a.name)
	}

	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}

	return filepath.Join(base, a.name)
}

// UserCache returns the users cache dir.
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

// UserData returns the users data dir.
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
