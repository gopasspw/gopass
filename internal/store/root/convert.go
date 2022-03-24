package root

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Convert will try to convert a given mount to a different set of
// backends.
func (r *Store) Convert(ctx context.Context, name string, cryptoBe backend.CryptoBackend, storageBe backend.StorageBackend, move bool) error {
	sub, err := r.GetSubStore(name)
	if err != nil {
		return fmt.Errorf("mount not found: %w", err)
	}

	debug.Log("converting %s to crypto: %s, rcs: %s, storage: %s", name, cryptoBe, storageBe)

	if err := sub.Convert(ctx, cryptoBe, storageBe, move); err != nil {
		return fmt.Errorf("failed to convert %q: %w", name, err)
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
