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
	cfg     *config.Config
	gpg     gpger
	mounts  map[string]*sub.Store
	path    string // path to the root store
	store   *sub.Store
	version string
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
		cfg:     cfg,
		gpg:     gpgcli.New(gpgcli.Config{}),
		mounts:  make(map[string]*sub.Store, len(cfg.Mounts)),
		path:    cfg.Path,
		version: cfg.Version,
	}

	// create the base store
	r.store = sub.New("", r.Path())

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

// Path returns the store path
func (r *Store) Path() string {
	return r.path
}

// Alias always returns an empty string
func (r *Store) Alias() string {
	return ""
}
