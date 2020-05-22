package root

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
)

// Crypto returns the crypto backend
func (r *Store) Crypto(ctx context.Context, name string) backend.Crypto {
	_, sub, _ := r.getStore(ctx, name)
	if !sub.Valid() {
		out.Debug(ctx, "Sub-Store not found for %s. Returning nil crypto backend", name)
		return nil
	}
	return sub.Crypto()
}
