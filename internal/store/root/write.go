package root

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/gopass"
)

// ValidateSecretName checks that name is a valid secret name for this store.
// It rejects empty names, names that match a mount point, and names that fail
// store.ValidateSecretName.
func (r *Store) ValidateSecretName(name string) error {
	sub, sn := r.getStore(name)
	if sn == "" {
		if sub != nil && sub.Alias() != "" {
			return fmt.Errorf("cannot create secret with the same name as mount point %q", sub.Alias())
		}

		return fmt.Errorf("invalid secret name %q: must not be empty", name)
	}

	return store.ValidateSecretName(sn)
}

// Set encodes and write the ciphertext of one entry to disk.
func (r *Store) Set(ctx context.Context, name string, sec gopass.Byter) error {
	if err := r.ValidateSecretName(name); err != nil {
		return err
	}

	sub, name := r.getStore(name)

	return sub.Set(ctx, name, sec)
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
