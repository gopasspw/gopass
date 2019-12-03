package openpgp

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
)

const (
	name = "openpgp"
)

func init() {
	backend.RegisterCrypto(backend.OpenPGP, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: %s (ALPHA)", name)
	return New(ctx)
}

func (l loader) String() string {
	return name
}
