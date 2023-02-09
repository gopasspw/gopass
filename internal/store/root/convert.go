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
		return fmt.Errorf("mount %q not found: %w", name, err)
	}

	debug.Log("converting %s to crypto: %s, storage: %s", name, cryptoBe, storageBe)

	if err := sub.Convert(ctx, cryptoBe, storageBe, move); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	if name == "" {
		debug.Log("success. updating root path to %s", sub.Path())

		return r.cfg.Set("", "mounts.path", sub.Path())
	}

	debug.Log("success. updating path for %s to %s", name, sub.Path())

	return r.cfg.Set("", "mounts."+name+".path", sub.Path())
}
