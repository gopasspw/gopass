package root

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/debug"
)

// Convert will try to convert a given mount to a different set of
// backends.
func (r *Store) Convert(ctx context.Context, name string, cryptoBe backend.CryptoBackend, rcsBe backend.RCSBackend, storageBe backend.StorageBackend, move bool) error {
	_, sub, err := r.GetSubStore(ctx, name)
	if err != nil {
		return err
	}
	debug.Log("converting %s to crypto: %s, rcs: %s, storage: %s", name, cryptoBe, rcsBe, storageBe)
	if err := sub.Convert(ctx, cryptoBe, rcsBe, storageBe, move); err != nil {
		return err
	}
	if name == "" {
		debug.Log("success. updating root path to %s", sub.Path())
		r.cfg.Path = sub.Path()
	} else {
		debug.Log("success. updating path for %s to %s", name, sub.Path())
		r.cfg.Mounts[name] = sub.Path()
	}

	return r.cfg.Save()
}
