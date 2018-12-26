package cli

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/out"
)

func init() {
	backend.RegisterCrypto(backend.GPGCLI, "gpgcli", &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: gpg-cli")
	return New(ctx, Config{
		Umask: fsutil.Umask(),
		Args:  GPGOpts(),
	})
}
