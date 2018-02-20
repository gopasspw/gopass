package root

import (
	"context"

	"github.com/justwatchcom/gopass/backend"
)

// Crypto returns the crypto backend
func (r *Store) Crypto(ctx context.Context, name string) backend.Crypto {
	_, sub, _ := r.getStore(ctx, name)
	return sub.Crypto()
}
