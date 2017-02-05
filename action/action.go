package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/gpg"
	"github.com/justwatchcom/gopass/password"
)

// Action knows everything to run gopass CLI actions
type Action struct {
	Name  string
	Store *password.RootStore
}

// New returns a new Action wrapper
func New(v string) *Action {
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	if gdb := os.Getenv("GOPASS_DEBUG"); gdb == "true" {
		gpg.Debug = true
	}
	pwDir := pwStoreDir("")

	// try to read config (if it exists)
	if cfg, err := newFromFile(configFile()); err == nil && cfg != nil {
		cfg.ImportFunc = askForKeyImport
		cfg.Version = v
		color.NoColor = cfg.NoColor

		return &Action{
			Name: name,
			Store: cfg,
		}
	}

	cfg, err := password.NewRootStore(pwDir)
	if err != nil {
		panic(err)
	}

	cfg.ImportFunc = askForKeyImport
	cfg.FsckFunc = askForConfirmation
	cfg.Version = v

	return &Action{
		Name:  name,
		Store: cfg,
	}
}

// newFromFile creates a new RootStore instance by unmarsahling a config file.
// If the file doesn't exist or fails to unmarshal an error is returned
func newFromFile(cf string) (*password.RootStore, error) {
	// deliberately using os.Stat here, a symlinked
	// config is OK
	if _, err := os.Stat(cf); err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadFile(cf)
	if err != nil {
		fmt.Printf("Error reading config from %s: %s\n", cf, err)
		return nil, err
	}
	cfg := &password.RootStore{}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		fmt.Printf("Error reading config from %s: %s\n", cf, err)
		return nil, err
	}
	return cfg, nil
}

// String implement fmt.Stringer
func (s *Action) String() string {
	return s.Store.String()
}

// pwStoreDir reads the password store dir from the environment
// or returns the default location ~/.password-store if the env is
// not set
func pwStoreDir(mount string) string {
	if mount != "" {
		return fsutil.CleanPath(filepath.Join(os.Getenv("HOME"), ".password-store-"+strings.Replace(mount, string(filepath.Separator), "-", -1)))
	}
	if d := os.Getenv("PASSWORD_STORE_DIR"); d != "" {
		return fsutil.CleanPath(d)
	}
	return os.Getenv("HOME") + "/.password-store"
}
