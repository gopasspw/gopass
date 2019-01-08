package plain

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
)

const (
	name = "plain"
)

func init() {
	backend.RegisterCrypto(backend.Plain, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: %s (NO ENCRYPTION)", name)
	return New(), nil
}

func (l loader) String() string {
	return name
}
