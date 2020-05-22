package cli

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const (
	name = "gpgcli"
)

func init() {
	backend.RegisterCrypto(backend.GPGCLI, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: %s", name)
	return New(ctx, Config{
		Umask: fsutil.Umask(),
		Args:  GPGOpts(),
	})
}

func (l loader) String() string {
	return name
}
