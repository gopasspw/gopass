package root

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// IsInitialized checks on disk if .gpg-id was generated and thus returns true.
func (r *Store) IsInitialized(ctx context.Context) (bool, error) {
	if r.store == nil {
		debug.Log("initializing store and possible sub-stores")
		if err := r.initialize(ctx); err != nil {
			return false, fmt.Errorf("failed to initialize stores: %w", err)
		}
	}
	debug.Log("root store is initialized")
	return r.store.IsInitialized(ctx), nil
}

// Init tries to initialize a new password store location matching the object
func (r *Store) Init(ctx context.Context, alias, path string, ids ...string) error {
	debug.Log("Instantiating new sub store %s at %s for %+v", alias, path, ids)
	if !backend.HasCryptoBackend(ctx) {
		ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)
	}
	if !backend.HasStorageBackend(ctx) {
		ctx = backend.WithStorageBackend(ctx, backend.GitFS)
	}
	sub, err := leaf.New(ctx, alias, path)
	if err != nil {
		return fmt.Errorf("failed to instantiate new sub store: %w", err)
	}
	if !r.store.IsInitialized(ctx) && alias == "" {
		r.store = sub
	}

	debug.Log("Initializing sub store at %s for %+v", path, ids)
	if err := sub.Init(ctx, path, ids...); err != nil {
		return fmt.Errorf("failed to initialize new sub store: %w", err)
	}

	if alias == "" {
		debug.Log("initialized root at %s", path)
		r.cfg.Path = path
	} else {
		debug.Log("mounted %s at %s", alias, path)
		r.cfg.Mounts[alias] = path
	}

	return nil
}

func (r *Store) initialize(ctx context.Context) error {
	// already initialized?
	if r.store != nil {
		return nil
	}

	// create the base store
	path := fsutil.CleanPath(r.cfg.Path)
	debug.Log("initialize - %s", path)
	s, err := leaf.New(ctx, "", path)
	if err != nil {
		return fmt.Errorf("failed to initialize the root store at %q: %w", r.cfg.Path, err)
	}
	debug.Log("Root Store initialized at %s", path)
	r.store = s

	// initialize all mounts
	for alias, path := range r.cfg.Mounts {
		path := fsutil.CleanPath(path)
		if err := r.addMount(ctx, alias, path); err != nil {
			out.Errorf(ctx, "Failed to initialize mount %s (%s). Ignoring: %s", alias, path, err)
			continue
		}
		debug.Log("Sub-Store mounted at %s from %s", alias, path)
	}

	// check for duplicate mounts
	if err := r.checkMounts(); err != nil {
		return fmt.Errorf("checking mounts failed: %w", err)
	}

	return nil
}
