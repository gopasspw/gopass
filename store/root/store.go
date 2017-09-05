package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/backend/gpg"
	gpgcli "github.com/justwatchcom/gopass/backend/gpg/cli"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/pkg/errors"
)

type gpger interface {
	FindPublicKeys(context.Context, ...string) (gpg.KeyList, error)
}

// Store is the public facing password store
type Store struct {
	askForMore  bool // context.TODO
	clipTimeout int  // clear clipboard after seconds // context.TODO
	gpg         gpger
	mounts      map[string]*sub.Store
	noColor     bool   // disable colors in output // context.TODO
	noConfirm   bool   // context.TODO
	noPager     bool   // context.TODO
	path        string // path to the root store
	safeContent bool   // avoid showing passwords in terminal // context.TODO
	store       *sub.Store
	version     string
}

// New creates a new store
func New(ctx context.Context, cfg *config.Config) (*Store, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.Path == "" {
		return nil, errors.Errorf("need path")
	}
	r := &Store{
		askForMore:  cfg.AskForMore,
		clipTimeout: cfg.ClipTimeout,
		gpg:         gpgcli.New(gpgcli.Config{}),
		mounts:      make(map[string]*sub.Store, len(cfg.Mounts)),
		noColor:     cfg.NoColor,
		noConfirm:   cfg.NoConfirm,
		noPager:     cfg.NoPager,
		path:        cfg.Path,
		safeContent: cfg.SafeContent,
		version:     cfg.Version,
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
		if err := r.addMount(ctx, alias, path); err != nil {
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
