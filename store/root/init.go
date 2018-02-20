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
		r.cfg.Root.CryptoBackend = backend.CryptoBackendName(backend.GetCryptoBackend(ctx))
		r.cfg.Root.SyncBackend = backend.SyncBackendName(backend.GetSyncBackend(ctx))
		r.cfg.Root.StoreBackend = backend.StoreBackendName(backend.GetStoreBackend(ctx))
	} else {
		if sc := r.cfg.Mounts[alias]; sc == nil {
			r.cfg.Mounts[alias] = &config.StoreConfig{}
		}
		r.cfg.Mounts[alias].CryptoBackend = backend.CryptoBackendName(backend.GetCryptoBackend(ctx))
		r.cfg.Mounts[alias].SyncBackend = backend.SyncBackendName(backend.GetSyncBackend(ctx))
		r.cfg.Mounts[alias].StoreBackend = backend.StoreBackendName(backend.GetStoreBackend(ctx))
	}
	return nil
}
