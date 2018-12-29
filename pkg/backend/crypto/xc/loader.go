package xc

import (
	"context"

	"github.com/gopasspw/gopass/pkg/agent/client"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
)

func init() {
	backend.RegisterCrypto(backend.XC, "xc", &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: xc (EXPERIMENTAL)")
	return New(ctxutil.GetConfigDir(ctx), client.GetClient(ctx))
}
