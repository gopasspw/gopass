// Package appdir implements a customized lookup pattern for application paths
// like config, cache and data dirs. On Linux this uses the XDG specification,
// on MacOS and Windows the platform defaults.
package appdir

import (
	"os"

	"github.com/gopasspw/gopass/pkg/debug"
)

// DefaultAppdir is the default appdir for gopass.
var DefaultAppdir = New("gopass")

// Appdir is a helper struct to generate paths for config, cache and data dirs.
type Appdir struct {
	// Name is used in the final path of the generated path.
	name string
}

// New returns a new Appdir for the given application name.
// The name is used to construct the paths to the application's
// directories.
func New(name string) *Appdir {
	return &Appdir{
		name: name,
	}
}

// Name returns the name of the appdir.
func (a *Appdir) Name() string {
	return a.name
}

// UserConfig returns the user's config dir for gopass.
// See a.UserConfig() for more details.
func UserConfig() string {
	return DefaultAppdir.UserConfig()
}

// UserCache returns the user's cache dir for gopass.
// See a.UserCache() for more details.
func UserCache() string {
	return DefaultAppdir.UserCache()
}

// UserData returns the user's data dir for gopass.
// See a.UserData() for more details.
func UserData() string {
	return DefaultAppdir.UserData()
}

// UserHome returns the user's home directory.
//
// If GOPASS_HOMEDIR is set it overrides the OS home dir lookup. This variable
// exists to allow hermetic test environments and CI runs to redirect all gopass
// data paths without modifying the real user profile. It should not be set in
// normal production use; doing so silently redirects every config, cache, and
// data path to the specified directory.
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
