package root

import (
	"context"
	"strings"

	"github.com/justwatchcom/gopass/store/secret"
)

// Set encodes and write the ciphertext of one entry to disk
func (r *Store) Set(ctx context.Context, name string, sec *secret.Secret) error {
	store := r.getStore(name)
	return store.Set(ctx, strings.TrimPrefix(name, store.Alias()), sec)
}
