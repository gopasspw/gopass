package root

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/leaf"

	"github.com/pkg/errors"
)

// Initialized checks on disk if .gpg-id was generated and thus returns true.
func (r *Store) Initialized(ctx context.Context) (bool, error) {
	if r.store == nil {
		out.Debug(ctx, "initializing store and possible sub-stores")
		if err := r.initialize(ctx); err != nil {
			return false, errors.Wrapf(err, "failed to initialized stores: %s", err)
		}
	}
	return r.store.Initialized(ctx), nil
}

// Init tries to initialize a new password store location matching the object
func (r *Store) Init(ctx context.Context, alias, path string, ids ...string) error {
	out.Debug(ctx, "Instantiating new sub store %s at %s for %+v", alias, path, ids)
	if !backend.HasCryptoBackend(ctx) {
		ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)
	}
	if !backend.HasRCSBackend(ctx) {
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	}
	if !backend.HasStorageBackend(ctx) {
		ctx = backend.WithStorageBackend(ctx, backend.FS)
	}
	sub, err := leaf.New(ctx, alias, path)
	if err != nil {
		return errors.Wrapf(err, "failed to instantiate new sub store: %s", err)
	}
	if !r.store.Initialized(ctx) && alias == "" {
		r.store = sub
	}

	out.Debug(ctx, "Initializing sub store at %s for %+v", path, ids)
	if err := sub.Init(ctx, path, ids...); err != nil {
		return errors.Wrapf(err, "failed to initialize new sub store: %s", err)
	}

	if alias == "" {
		out.Debug(ctx, "initialized root at %s", path)
		r.cfg.Path = path
	} else {
		out.Debug(ctx, "mounted %s at %s", alias, path)
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
	out.Debug(ctx, "initialize - %s", r.cfg.Path)
	s, err := leaf.New(ctx, "", r.cfg.Path)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize the root store at '%s': %s", r.cfg.Path, err)
	}
	out.Debug(ctx, "Root Store initialized at %s", r.cfg.Path)
	r.store = s

	// initialize all mounts
	for alias, path := range r.cfg.Mounts {
		if err := r.addMount(ctx, alias, path); err != nil {
			out.Error(ctx, "Failed to initialize mount %s (%s). Ignoring: %s", alias, path, err)
			continue
		}
		out.Debug(ctx, "Sub-Store mounted at %s from %s", alias, path)
	}

	// check for duplicate mounts
	if err := r.checkMounts(); err != nil {
		return errors.Errorf("checking mounts failed: %s", err)
	}

	return nil
}
