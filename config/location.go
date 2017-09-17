package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/justwatchcom/gopass/utils/fsutil"
	homedir "github.com/mitchellh/go-homedir"
)

// Homedir returns the users home dir or an empty string if the lookup fails
func Homedir() string {
	hd, err := homedir.Dir()
	if err != nil {
		if debug {
			fmt.Printf("[DEBUG] Failed to get homedir: %s\n", err)
		}
		return ""
	}
	return hd
}

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

	return filepath.Join(Homedir(), ".config", "gopass", "config.yml")
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
	l = append(l, filepath.Join(Homedir(), ".config", "gopass", "config.yml"))
	l = append(l, filepath.Join(Homedir(), ".gopass.yml"))
	return l
}

// PwStoreDir reads the password store dir from the environment
// or returns the default location ~/.password-store if the env is
// not set
func PwStoreDir(mount string) string {
	if mount != "" {
		return fsutil.CleanPath(filepath.Join(Homedir(), ".password-store-"+strings.Replace(mount, string(filepath.Separator), "-", -1)))
	}
	if d := os.Getenv("PASSWORD_STORE_DIR"); d != "" {
		return fsutil.CleanPath(d)
	}
	return filepath.Join(Homedir(), ".password-store")
}
