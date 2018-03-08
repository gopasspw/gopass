package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// Store is the public facing password store
type Store struct {
	cfg     *config.Config
	mounts  map[string]*sub.Store
	url     *backend.URL // url of the root store
	store   *sub.Store
	version string
}

// New creates a new store
func New(ctx context.Context, cfg *config.Config) (*Store, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.Root != nil && (cfg.Root.Path == nil || cfg.Root.Path.Path == "") {
		return nil, errors.Errorf("need path")
	}
	r := &Store{
		cfg:     cfg,
		mounts:  make(map[string]*sub.Store, len(cfg.Mounts)),
		url:     cfg.Root.Path,
		version: cfg.Version,
	}

	// create the base store
	if !backend.HasCryptoBackend(ctx) {
		ctx = backend.WithCryptoBackend(ctx, cfg.Root.Path.Crypto)
	}
	if !backend.HasSyncBackend(ctx) {
		ctx = backend.WithSyncBackend(ctx, cfg.Root.Path.Sync)
	}
	s, err := sub.New(ctx, "", r.Path(), config.Directory())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize the root store at '%s': %s", r.Path(), err)
	}
	r.store = s

	// initialize all mounts
	for alias, sc := range cfg.Mounts {
		if err := r.addMount(ctx, alias, sc.Path.String(), sc); err != nil {
			out.Red(ctx, "Failed to initialize mount %s (%s). Ignoring: %s", alias, sc.Path.String(), err)
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
func (r *Store) Exists(ctx context.Context, name string) bool {
	_, store, name := r.getStore(ctx, name)
	return store.Exists(ctx, name)
}

// IsDir checks if a given key is actually a folder
func (r *Store) IsDir(ctx context.Context, name string) bool {
	_, store, name := r.getStore(ctx, name)
	return store.IsDir(ctx, name)
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
	return r.url.Path
}

// Alias always returns an empty string
func (r *Store) Alias() string {
	return ""
}

// Store returns the storage backend for the given mount point
func (r *Store) Store(ctx context.Context, name string) backend.Store {
	_, sub, _ := r.getStore(ctx, name)
	return sub.Store()
}
