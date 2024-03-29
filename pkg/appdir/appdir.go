// Package appdir implements a customized lookup pattern for application paths
// like config, cache and data dirs. On Linux this uses the XDG specification,
// on MacOS and Windows the platform defaults.
package appdir

import (
	"os"

	"github.com/gopasspw/gopass/pkg/debug"
)

var DefaultAppdir = New("gopass")

// Appdir is a helper struct to generate paths for config, cache and data dirs.
type Appdir struct {
	// Name is used in the final path of the generated path.
	name string
}

// New returns a new Appdir.
func New(name string) *Appdir {
	return &Appdir{
		name: name,
	}
}

// Name returns the name of the appdir.
func (a *Appdir) Name() string {
	return a.name
}

// UserConfig returns the users config dir.
func UserConfig() string {
	return DefaultAppdir.UserConfig()
}

// UserCache returns the users cache dir.
func UserCache() string {
	return DefaultAppdir.UserCache()
}

// UserData returns the users data dir.
func UserData() string {
	return DefaultAppdir.UserData()
}

// UserHome returns the users home dir.
func UserHome() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return hd
	}

	uhd, err := os.UserHomeDir()
	if err != nil {
		debug.Log("failed to detect user home dir: %s", err)

		return ""
	}

	return uhd
}
