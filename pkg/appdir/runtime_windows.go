package appdir

import (
	"os"
	"path/filepath"
)

// UserRuntime returns the users runtime dir
func (a *Appdir) UserRuntime() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return filepath.Join(hd, ".run")
	}

	return filepath.Join(os.Getenv("LOCALAPPDATA"), a.name)
}

// UserRuntime returns the users runtime dir.
func UserRuntime() string {
	return DefaultAppdir.UserRuntime()
}
