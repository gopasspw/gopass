package appdir

import (
	"os"
	"path/filepath"
)

// UserConfig returns the users config dir
func UserConfig() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".config", "gopass")
	}

	return filepath.Join(os.Getenv("APPDATA"), "gopass")
}

// UserCache returns the users cache dir
func UserCache() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".cache", "gopass")
	}

	return filepath.Join(os.Getenv("LOCALAPPDATA"), "gopass")
}

// UserData returns the users data dir
func UserData() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".local", "share", "gopass")
	}
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "gopass")
}
