package appdir

import (
	"os"
	"path/filepath"
)

// UserConfig returns the users config dir
func UserConfig() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".config", Name)
	}

	return filepath.Join(os.Getenv("APPDATA"), Name)
}

// UserCache returns the users cache dir
func UserCache() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".cache", Name)
	}

	return filepath.Join(os.Getenv("LOCALAPPDATA"), Name)
}

// UserData returns the users data dir
func UserData() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".local", "share", Name)
	}
	return filepath.Join(os.Getenv("LOCALAPPDATA"), Name)
}
