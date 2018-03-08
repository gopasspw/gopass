package root

import (
	"context"

	"github.com/justwatchcom/gopass/store/secret"
)

// Get returns the plaintext of a single key
func (r *Store) Get(ctx context.Context, name string) (*secret.Secret, error) {
	// forward to substore
	ctx, store, name := r.getStore(ctx, name)
	return store.Get(ctx, name)
}

// GetContext returns the plaintext and the context of a single key
func (r *Store) GetContext(ctx context.Context, name string) (*secret.Secret, context.Context, error) {
	// forward to substore
	ctx, store, name := r.getStore(ctx, name)
	sec, err := store.Get(ctx, name)
	return sec, ctx, err
}
