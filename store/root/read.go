package root

import (
	"context"
	"strings"

	"github.com/justwatchcom/gopass/store/secret"
)

// Get returns the plaintext of a single key
func (r *Store) Get(ctx context.Context, name string) (*secret.Secret, error) {
	// forward to substore
	store := r.getStore(name)
	return store.Get(ctx, strings.TrimPrefix(name, store.Alias()))
}
