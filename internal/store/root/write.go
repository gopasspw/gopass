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

// CheckImport checks the storage state of a passkey import target.
func (r *Store) CheckImport(ctx context.Context, name string, first bool) error {
	sub, path := r.getStore(name)
	if sub == nil || path == "" {
		return fmt.Errorf("no store for import target %q", name)
	}

	return sub.CheckImport(ctx, path, first)
}
