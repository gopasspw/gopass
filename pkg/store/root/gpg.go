package root

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/backend"
)

// Crypto returns the crypto backend
func (r *Store) Crypto(ctx context.Context, name string) backend.Crypto {
	_, sub, _ := r.getStore(ctx, name)
	if !sub.Valid() {
		return nil
	}
	return sub.Crypto()
}
