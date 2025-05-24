package root

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/gopass"
)

// Get returns the plaintext of a single key.
func (r *Store) Get(ctx context.Context, name string) (gopass.Secret, error) {
	store, name := r.getStore(name)

	sec, err := store.Get(ctx, name)
	if err != nil {
		return sec, err
	}

	if ref, ok := sec.Ref(); ok {
		refSec, err := store.Get(ctx, ref)
		if err != nil {
			return sec, fmt.Errorf("failed to read reference %s by %s: %w", ref, name, err)
		}

		sec.SetPassword(refSec.Password())
	}

	return sec, nil
}
