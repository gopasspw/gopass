package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/justwatchcom/gopass/utils/fsutil"
)

// configLocation returns the location of the config file
// (a YAML file that contains values such as the path to the password store)
func configLocation() string {
	// First, check for the "GOPASS_CONFIG" environment variable
	if cf := os.Getenv("GOPASS_CONFIG"); cf != "" {
		return cf
	}

	// Second, check for the "XDG_CONFIG_HOME" environment variable
	// (which is part of the XDG Base Directory Specification for Linux and
	// other Unix-like operating sytstems)
	if xch := os.Getenv("XDG_CONFIG_HOME"); xch != "" {
		return filepath.Join(xch, "gopass", "config.yml")
	}

	// Third, check to see if we are running on a Windows platform
	// We can check for platform via the "runtime.GOOS" variable:
	// https://stackoverflow.com/questions/19847594/how-to-reliably-detect-os-platform-in-go
	if runtime.GOOS == "windows" {
		// Windows uses the "userprofile" environment variable instead of "HOME":
		// https://stackoverflow.com/questions/9228950/what-is-the-alternative-for-users-home-directory-on-windows-command-prompt
		return filepath.Join(os.Getenv("userprofile"), ".config", "gopass", "config.yml")
	}

	// Default to using the "HOME" environment variable present on most Linux &
	// OS X systems
	return filepath.Join(os.Getenv("HOME"), ".config", "gopass", "config.yml")
}

// configLocations returns the possible locations of gopass config files,
// in decreasing priority
func configLocations() []string {
	l := []string{}
	if cf := os.Getenv("GOPASS_CONFIG"); cf != "" {
		l = append(l, cf)
	}
	if xch := os.Getenv("XDG_CONFIG_HOME"); xch != "" {
		l = append(l, filepath.Join(xch, "gopass", "config.yml"))
	}
	l = append(l, filepath.Join(os.Getenv("HOME"), ".config", "gopass", "config.yml"))
	l = append(l, filepath.Join(os.Getenv("HOME"), ".gopass.yml"))
	return l
}

// PwStoreDir reads the password store dir from the environment
// or returns the default location ~/.password-store if the env is
// not set
func PwStoreDir(mount string) string {
	if mount != "" {
		return fsutil.CleanPath(filepath.Join(os.Getenv("HOME"), ".password-store-"+strings.Replace(mount, string(filepath.Separator), "-", -1)))
	}
	if d := os.Getenv("PASSWORD_STORE_DIR"); d != "" {
		return fsutil.CleanPath(d)
	}
	return os.Getenv("HOME") + "/.password-store"
}
