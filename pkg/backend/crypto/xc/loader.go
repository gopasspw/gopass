package xc

import (
	"context"

	"github.com/gopasspw/gopass/pkg/agent/client"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
)

const (
	name = "xc"
)

func Init() {
	backend.RegisterCrypto(backend.XC, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: %s (EXPERIMENTAL)", name)
	return New(ctxutil.GetConfigDir(ctx), client.GetClient(ctx))
}

func (l loader) String() string {
	return name
}
