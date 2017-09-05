package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/gpg"
	gpgcli "github.com/justwatchcom/gopass/gpg/cli"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/pkg/errors"
)

type gpger interface {
	FindPublicKeys(context.Context, ...string) (gpg.KeyList, error)
}

// Store is the public facing password store
type Store struct {
	askForMore  bool
	autoImport  bool
	autoSync    bool // push to git remote after commit
	clipTimeout int  // clear clipboard after seconds
	fsckFunc    store.FsckCallback
	gpg         gpger
	importFunc  store.ImportCallback
	mounts      map[string]*sub.Store
	noColor     bool // disable colors in output
	noConfirm   bool
	noPager     bool
	path        string // path to the root store
	safeContent bool   // avoid showing passwords in terminal
	store       *sub.Store
	version     string
}

// New creates a new store
func New(cfg *config.Config) (*Store, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.Path == "" {
		return nil, errors.Errorf("need path")
	}
	r := &Store{
		askForMore:  cfg.AskForMore,
		autoImport:  cfg.AutoImport,
		autoSync:    cfg.AutoSync,
		clipTimeout: cfg.ClipTimeout,
		fsckFunc:    cfg.FsckFunc,
		gpg: gpgcli.New(gpgcli.Config{
			AlwaysTrust: true,
		}),
		importFunc:  cfg.ImportFunc,
		mounts:      make(map[string]*sub.Store, len(cfg.Mounts)),
		noColor:     cfg.NoColor,
		noConfirm:   cfg.NoConfirm,
		noPager:     cfg.NoPager,
		path:        cfg.Path,
		safeContent: cfg.SafeContent,
		version:     cfg.Version,
	}

	if r.autoImport {
		r.importFunc = nil
	}

	// create the base store
	subCfg := r.Config()
	subCfg.Path = fsutil.CleanPath(r.Path())
	s, err := sub.New("", subCfg)
	if err != nil {
		return nil, err
	}

	r.store = s

	// initialize all mounts
	for alias, path := range cfg.Mounts {
		path = fsutil.CleanPath(path)
		if err := r.addMount(context.TODO(), alias, path); err != nil {
			fmt.Printf("Failed to initialized mount %s (%s): %s. Ignoring\n", alias, path, err)
			continue
		}
	}

	// check for duplicate mounts
	if err := r.checkMounts(); err != nil {
		return nil, errors.Errorf("checking mounts failed: %s", err)
	}

	// set some defaults
	if r.clipTimeout < 1 {
		r.clipTimeout = 45
	}

	return r, nil
}

// Exists checks the existence of a single entry
func (r *Store) Exists(name string) bool {
	store := r.getStore(name)
	return store.Exists(strings.TrimPrefix(name, store.Alias()))
}

// IsDir checks if a given key is actually a folder
func (r *Store) IsDir(name string) bool {
	store := r.getStore(name)
	return store.IsDir(strings.TrimPrefix(name, store.Alias()))
}

func (r *Store) String() string {
	ms := make([]string, 0, len(r.mounts))
	for alias, sub := range r.mounts {
		ms = append(ms, alias+"="+sub.String())
	}
	return fmt.Sprintf("Store(Path: %s, Mounts: %+v)", r.store.Path(), strings.Join(ms, ","))
}
