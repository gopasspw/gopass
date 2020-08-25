// Package appdir implements a customized lookup pattern for application paths
// like config, cache and data dirs. On Linux this uses the XDG specification,
// on MacOS and Windows the platform defaults.
package appdir

import (
	"os"

	"github.com/gopasspw/gopass/internal/debug"
)

// UserHome returns the users home dir
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
