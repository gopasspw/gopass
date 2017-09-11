package root

import (
	"context"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/store/secret"
)

// Set encodes and write the ciphertext of one entry to disk
func (r *Store) Set(ctx context.Context, name string, sec *secret.Secret, reason string) error {
	store := r.getStore(name)
	return store.Set(ctx, strings.TrimPrefix(name, store.Alias()), sec, reason)
}

// SetConfirm calls Set with confirmation callback
func (r *Store) SetConfirm(ctx context.Context, name string, sec *secret.Secret, reason string, cb store.RecipientCallback) error {
	store := r.getStore(name)
	return store.SetConfirm(ctx, strings.TrimPrefix(name, store.Alias()), sec, reason, cb)
}
