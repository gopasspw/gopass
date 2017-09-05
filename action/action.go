package action

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/backend/gpg"
	gpgcli "github.com/justwatchcom/gopass/backend/gpg/cli"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/root"
)

type gpger interface {
	FindPublicKeys(context.Context, ...string) (gpg.KeyList, error)
	FindPrivateKeys(context.Context, ...string) (gpg.KeyList, error)
	ListPublicKeys(context.Context) (gpg.KeyList, error)
	ListPrivateKeys(context.Context) (gpg.KeyList, error)
}

// Action knows everything to run gopass CLI actions
type Action struct {
	Name    string
	Store   *root.Store
	gpg     gpger
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
	}
	cfg.ImportFunc = act.askForKeyImport
	cfg.FsckFunc = act.askForConfirmation

	store, err := root.New(cfg)
	if err != nil {
		panic(err)
	}
	act.Store = store

	act.gpg = gpgcli.New(gpgcli.Config{
		AlwaysTrust: true,
	})

	return act
}

// String implement fmt.Stringer
func (s *Action) String() string {
	return s.Store.String()
}
