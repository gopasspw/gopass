package root

import (
	"context"

	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
)

// Initialized checks on disk if .gpg-id was generated and thus returns true.
func (r *Store) Initialized(ctx context.Context) bool {
	return r.store.Initialized(ctx)
}

// Init tries to initialize a new password store location matching the object
func (r *Store) Init(ctx context.Context, alias, path string, ids ...string) error {
	sub, err := sub.New(ctx, alias, path, config.Directory())
	if err != nil {
		return err
	}
	if !r.store.Initialized(ctx) && alias == "" {
		r.store = sub
	}

	if err := sub.Init(ctx, path, ids...); err != nil {
		return err
	}
	if alias == "" {
		if r.cfg.Root.Path == nil {
			r.cfg.Root.Path = backend.FromPath(path)
		}
		r.cfg.Root.Path.Crypto = backend.GetCryptoBackend(ctx)
		r.cfg.Root.Path.Sync = backend.GetSyncBackend(ctx)
		r.cfg.Root.Path.Store = backend.GetStoreBackend(ctx)
	} else {
		if sc := r.cfg.Mounts[alias]; sc == nil {
			r.cfg.Mounts[alias] = &config.StoreConfig{}
		}
		if r.cfg.Mounts[alias].Path == nil {
			r.cfg.Mounts[alias].Path = backend.FromPath(path)
		}
		r.cfg.Mounts[alias].Path.Crypto = backend.GetCryptoBackend(ctx)
		r.cfg.Mounts[alias].Path.Sync = backend.GetSyncBackend(ctx)
		r.cfg.Mounts[alias].Path.Store = backend.GetStoreBackend(ctx)
	}
	return nil
}
