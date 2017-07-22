package action

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/gpg"
	"github.com/justwatchcom/gopass/store/root"
)

// Action knows everything to run gopass CLI actions
type Action struct {
	Name    string
	Store   *root.Store
	gpg     *gpg.GPG
	isTerm  bool
	version semver.Version
}

// New returns a new Action wrapper
func New(sv semver.Version) *Action {
	name := "gopass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	// try to read config (if it exists)
	cfg := config.Load()
	// only update version field in config, if it's older than this build
	csv, err := semver.Parse(cfg.Version)
	if err != nil || csv.LT(sv) {
		cfg.Version = sv.String()
		if err := cfg.Save(); err != nil {
			fmt.Println(color.RedString("Failed to save config: %s", err))
		}
	}

	act := &Action{
		Name:    name,
		version: sv,
		isTerm:  true,
	}
	cfg.ImportFunc = act.askForKeyImport
	cfg.FsckFunc = act.askForConfirmation

	// debug flag
	if gdb := os.Getenv("GOPASS_DEBUG"); gdb == "true" {
		cfg.Debug = true
	}

	// need this override for our integration tests
	if nc := os.Getenv("GOPASS_NOCOLOR"); nc == "true" {
		cfg.NoColor = true
		color.NoColor = true
	}

	// only emit color codes when stdout is a terminal
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		cfg.NoColor = true
		color.NoColor = true
		cfg.ImportFunc = nil
		cfg.FsckFunc = nil
		act.isTerm = false
	}

	// reading from stdin?
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice == 0 {
		cfg.ImportFunc = nil
		cfg.FsckFunc = nil
		act.isTerm = false
	}

	store, err := root.New(cfg)
	if err != nil {
		panic(err)
	}
	act.Store = store

	act.gpg = gpg.New(gpg.Config{
		Debug:       cfg.Debug,
		AlwaysTrust: cfg.AlwaysTrust,
	})

	return act
}

// String implement fmt.Stringer
func (s *Action) String() string {
	return s.Store.String()
}
