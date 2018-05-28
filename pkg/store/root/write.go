package root

import (
	"context"

	"github.com/gopasspw/gopass/pkg/store"
)

// Set encodes and write the ciphertext of one entry to disk
func (r *Store) Set(ctx context.Context, name string, sec store.Secret) error {
	ctx, store, name := r.getStore(ctx, name)
	return store.Set(ctx, name, sec)
}

// SetContext encodes and write the ciphertext of one entry to disk and propagate the context
func (r *Store) SetContext(ctx context.Context, name string, sec store.Secret) (context.Context, error) {
	ctx, store, name := r.getStore(ctx, name)
	return ctx, store.Set(ctx, name, sec)
}
