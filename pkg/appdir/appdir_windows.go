package appdir

import (
	"os"
	"path/filepath"
)

// UserConfig returns the users config dir
func (a *Appdir) UserConfig() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".config", a.name)
	}

	return filepath.Join(os.Getenv("APPDATA"), a.name)
}

// UserCache returns the users cache dir
func (a *Appdir) UserCache() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".cache", a.ame)
	}

	return filepath.Join(os.Getenv("LOCALAPPDATA"), a.ame)
}

// UserData returns the users data dir
func (a *Appdir) UserData() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".local", "share", a.ame)
	}
	return filepath.Join(os.Getenv("LOCALAPPDATA"), a.ame)
}
