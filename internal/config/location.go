package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"

	homedir "github.com/mitchellh/go-homedir"
)

// Homedir returns the users home dir or an empty string if the lookup fails
func Homedir() string {
	if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
		return hd
	}
	hd, err := homedir.Dir()
	if err != nil {
		debug.Log("Failed to get homedir: %s\n", err)
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
	return filepath.Join(appdir.UserConfig(), "config.yml")
}

// configLocations returns the possible locations of gopass config files,
// in decreasing priority
func configLocations() []string {
	l := []string{}
	if cf := os.Getenv("GOPASS_CONFIG"); cf != "" {
		l = append(l, cf)
	}
	l = append(l, filepath.Join(appdir.UserConfig(), "config.yml"))
	l = append(l, filepath.Join(Homedir(), ".config", "gopass", "config.yml"))
	l = append(l, filepath.Join(Homedir(), ".gopass.yml"))
	return l
}

// PwStoreDir reads the password store dir from the environment
// or returns the default location if the env is not set
func PwStoreDir(mount string) string {
	if mount != "" {
		cleanName := strings.Replace(mount, string(filepath.Separator), "-", -1)
		return fsutil.CleanPath(filepath.Join(appdir.UserData(), "stores", cleanName))
	}
	// TODO(2.x): PASSWORD_STORE_DIR support is deprecated
	if d := os.Getenv("PASSWORD_STORE_DIR"); d != "" {
		return fsutil.CleanPath(d)
	}
	if ld := filepath.Join(appdir.UserHome(), ".password-store"); fsutil.IsDir(ld) {
		debug.Log("re-using existing legacy dir for root store: %s", ld)
		return ld
	}
	return fsutil.CleanPath(filepath.Join(appdir.UserData(), "stores", "root"))
}

// Directory returns the configuration directory for the gopass config file
func Directory() string {
	return filepath.Dir(configLocation())
}
