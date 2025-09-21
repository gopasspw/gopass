package appdir

import (
	"os"
	"path/filepath"
)

// UserConfig returns the user's config directory.
// It uses the APPDATA environment variable on Windows.
// The GOPASS_HOMEDIR environment variable can be used to override the base path.
func (a *Appdir) UserConfig() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".config", a.name)
	}

	return filepath.Join(os.Getenv("APPDATA"), a.name)
}

// UserCache returns the user's cache directory.
// It uses the LOCALAPPDATA environment variable on Windows.
// The GOPASS_HOMEDIR environment variable can be used to override the base path.
func (a *Appdir) UserCache() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".cache", a.name)
	}

	return filepath.Join(os.Getenv("LOCALAPPDATA"), a.name)
}

// UserData returns the user's data directory.
// It uses the LOCALAPPDATA environment variable on Windows.
// The GOPASS_HOMEDIR environment variable can be used to override the base path.
func (a *Appdir) UserData() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".local", "share", a.name)
	}
	return filepath.Join(os.Getenv("LOCALAPPDATA"), a.name)
}
