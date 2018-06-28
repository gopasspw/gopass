package root

import (
	"context"

	"github.com/gopasspw/gopass/pkg/agent/client"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/sub"

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
	// parse backend URL
	pathURL, err := backend.ParseURL(path)
	if err != nil {
		return errors.Wrapf(err, "failed to parse backend URL '%s': %s", path, err)
	}
	sub, err := sub.New(ctx, r.cfg, alias, pathURL, r.cfg.Directory(), r.agent)
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

	return r.initConfig(ctx, alias, path)
}

func (r *Store) initConfig(ctx context.Context, alias, path string) error {
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
		return nil
	}

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
		bu, err := backend.ParseURL(r.url.String())
		if err != nil {
			return errors.Wrapf(err, "failed to parse backend URL '%s': %s", r.url.String(), err)
		}
		s, err := sub.New(ctx, r.cfg, "", bu, r.cfg.Directory(), r.agent)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize the root store at '%s': %s", r.url.String(), err)
		}
		out.Debug(ctx, "Root Store initialized with URL %s", r.url.String())
		r.store = s
	}

	// initialize all mounts
	for alias, sc := range r.cfg.Mounts {
		if err := r.addMount(ctx, alias, sc.Path.String(), sc); err != nil {
			out.Error(ctx, "Failed to initialize mount %s (%s). Ignoring: %s", alias, sc.Path.String(), err)
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
