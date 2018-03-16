package root

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/agent/client"
	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/config"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store/sub"
	"github.com/pkg/errors"
)

// Initialized checks on disk if .gpg-id was generated and thus returns true.
func (r *Store) Initialized(ctx context.Context) bool {
	if r.store == nil {
		out.Debug(ctx, "initializing store and possible sub-stores")
		if err := r.initialize(ctx); err != nil {
			out.Red(ctx, "Faild to initialize stores: %s", err)
			return false
		}
	}
	return r.store.Initialized(ctx)
}

// Init tries to initialize a new password store location matching the object
func (r *Store) Init(ctx context.Context, alias, path string, ids ...string) error {
	out.Debug(ctx, "Instantiating new sub store %s at %s for %+v", alias, path, ids)
	sub, err := sub.New(ctx, alias, path, r.cfg.Directory(), r.agent)
	if err != nil {
		return err
	}
	if !r.store.Initialized(ctx) && alias == "" {
		r.store = sub
	}

	out.Debug(ctx, "Initializing sub store at %s for %+v", path, ids)
	if err := sub.Init(ctx, path, ids...); err != nil {
		return err
	}
	if alias == "" {
		if r.cfg.Root.Path == nil {
			r.cfg.Root.Path = backend.FromPath(path)
		}
		if backend.HasCryptoBackend(ctx) {
			r.cfg.Root.Path.Crypto = backend.GetCryptoBackend(ctx)
		}
		if backend.HasRCSBackend(ctx) {
			r.cfg.Root.Path.RCS = backend.GetRCSBackend(ctx)
		}
		if backend.HasStorageBackend(ctx) {
			r.cfg.Root.Path.Storage = backend.GetStorageBackend(ctx)
		}
	} else {
		if sc := r.cfg.Mounts[alias]; sc == nil {
			r.cfg.Mounts[alias] = &config.StoreConfig{}
		}
		if r.cfg.Mounts[alias].Path == nil {
			r.cfg.Mounts[alias].Path = backend.FromPath(path)
		}
		if backend.HasCryptoBackend(ctx) {
			r.cfg.Mounts[alias].Path.Crypto = backend.GetCryptoBackend(ctx)
		}
		if backend.HasRCSBackend(ctx) {
			r.cfg.Mounts[alias].Path.RCS = backend.GetRCSBackend(ctx)
		}
		if backend.HasStorageBackend(ctx) {
			r.cfg.Mounts[alias].Path.Storage = backend.GetStorageBackend(ctx)
		}
	}
	return nil
}

func (r *Store) initialize(ctx context.Context) error {
	// already initialized?
	if r.store != nil {
		return nil
	}

	// init agent client
	r.agent = client.New(config.Directory())

	// create the base store
	{
		// capture ctx to limit effect on the next sub.New call and to not
		// propagate it's effects to the mounts below
		ctx := ctx
		if !backend.HasCryptoBackend(ctx) {
			ctx = backend.WithCryptoBackend(ctx, r.cfg.Root.Path.Crypto)
		}
		if !backend.HasRCSBackend(ctx) {
			ctx = backend.WithRCSBackend(ctx, r.cfg.Root.Path.RCS)
		}
		if !backend.HasStorageBackend(ctx) {
			ctx = backend.WithStorageBackend(ctx, r.cfg.Root.Path.Storage)
		}
		s, err := sub.New(ctx, "", r.url.String(), r.cfg.Directory(), r.agent)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize the root store at '%s': %s", r.url.String(), err)
		}
		out.Debug(ctx, "Root Store initialized with URL %s", r.url.String())
		r.store = s
	}

	// initialize all mounts
	for alias, sc := range r.cfg.Mounts {
		if err := r.addMount(ctx, alias, sc.Path.String(), sc); err != nil {
			out.Red(ctx, "Failed to initialize mount %s (%s). Ignoring: %s", alias, sc.Path.String(), err)
			continue
		}
		out.Debug(ctx, "Sub-Store mounted at %s from %s", alias, sc.Path.String())
	}

	// check for duplicate mounts
	if err := r.checkMounts(); err != nil {
		return errors.Errorf("checking mounts failed: %s", err)
	}

	return nil
}
