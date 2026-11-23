package root

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/gopass"
)

// Set encodes and write the ciphertext of one entry to disk.
func (r *Store) Set(ctx context.Context, name string, sec gopass.Byter) error {
	store, name := r.getStore(name)

	return store.Set(ctx, name, sec)
}

// SameRecipientScope reports whether writes to two names use the same store and recipient file.
func (r *Store) SameRecipientScope(ctx context.Context, a, b string) bool {
	subA, a := r.getStore(a)
	subB, b := r.getStore(b)

	return subA != nil && subA == subB && subA.SameRecipientScope(ctx, a, b)
}

// StageUncommitted stages earlier writes that did not get committed in a mount.
func (r *Store) StageUncommitted(ctx context.Context, mount string, names ...string) (bool, error) {
	sub, err := r.GetSubStore(mount)
	if err != nil {
		return false, err
	}
	relative := make([]string, 0, len(names))
	for _, name := range names {
		target, path := r.getStore(name)
		if sub == nil || sub != target || path == "" {
			return false, fmt.Errorf("secret %q is not in mount %q", name, mount)
		}
		relative = append(relative, path)
	}

	return sub.StageUncommitted(ctx, relative...)
}
