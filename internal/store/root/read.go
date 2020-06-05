package root

import (
	"context"

	"github.com/gopasspw/gopass/internal/store"
)

// Get returns the plaintext of a single key
func (r *Store) Get(ctx context.Context, name string) (store.Secret, error) {
	// forward to substore
	ctx, store, name := r.getStore(ctx, name)
	return store.Get(ctx, name)
}
