package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/justwatchcom/gopass/utils/fsutil"
)

// configLocation returns the location of the config file. Either reading from
// GOPASS_CONFIG or using the default location (~/.gopass.yml)
func configLocation() string {
	if cf := os.Getenv("GOPASS_CONFIG"); cf != "" {
		return cf
	}
	if xch := os.Getenv("XDG_CONFIG_HOME"); xch != "" {
		return filepath.Join(xch, "gopass", "config.yml")
	}
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
