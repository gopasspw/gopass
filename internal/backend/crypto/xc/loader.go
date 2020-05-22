package xc

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
)

const (
	name = "xc"
)

func init() {
	backend.RegisterCrypto(backend.XC, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: %s (EXPERIMENTAL)", name)
	return New(ctxutil.GetConfigDir(ctx), nil)
}

func (l loader) String() string {
	return name
}
