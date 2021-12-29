package root

import (
	"context"

	"github.com/gopasspw/gopass/pkg/gopass"
)

// Get returns the plaintext of a single key.
func (r *Store) Get(ctx context.Context, name string) (gopass.Secret, error) {
	store, name := r.getStore(name)
	return store.Get(ctx, name)
}
