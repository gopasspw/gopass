package root

import (
	"context"

	"github.com/justwatchcom/gopass/store/secret"
)

// Set encodes and write the ciphertext of one entry to disk
func (r *Store) Set(ctx context.Context, name string, sec *secret.Secret) error {
	ctx, store, name := r.getStore(ctx, name)
	return store.Set(ctx, name, sec)
}
