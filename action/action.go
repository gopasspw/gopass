package action

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/gpg"
	"github.com/justwatchcom/gopass/store/root"
)

// Action knows everything to run gopass CLI actions
type Action struct {
	Name  string
	Store *root.Store
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

	// try to read config (if it exists)
	if cfg, err := config.Load(); err == nil && cfg != nil {
		cfg.ImportFunc = askForKeyImport
		cfg.FsckFunc = askForConfirmation
		cfg.Version = v
		color.NoColor = cfg.NoColor
		// need this override for our integration tests
		if nc := os.Getenv("GOPASS_NOCOLOR"); nc == "true" {
			color.NoColor = true
		}
		// only emit color codes when stdout is a terminal
		if !terminal.IsTerminal(int(os.Stdout.Fd())) {
			color.NoColor = true
		}
		store, err := root.New(cfg)
		if err != nil {
			panic(err)
		}
		return &Action{
			Name:  name,
			Store: store,
		}
	}

	cfg := config.New()
	cfg.Path = pwStoreDir("")
	cfg.ImportFunc = askForKeyImport
	cfg.FsckFunc = askForConfirmation
	rs, err := root.New(cfg)
	if err != nil {
		panic(err)
	}

	// need this override for our integration tests
	if nc := os.Getenv("GOPASS_NOCOLOR"); nc == "true" {
		color.NoColor = true
	}
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		color.NoColor = true
	}

	return &Action{
		Name:  name,
		Store: rs,
	}
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
