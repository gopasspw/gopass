package plain

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
)

func init() {
	backend.RegisterCrypto(backend.Plain, "plain", &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: plain (NO ENCRYPTION)")
	return New(), nil
}
