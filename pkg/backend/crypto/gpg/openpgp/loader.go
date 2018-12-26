package openpgp

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
)

func init() {
	backend.RegisterCrypto(backend.OpenPGP, "openpgp", &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: openpgp (ALPHA)")
	return New(ctx)
}
