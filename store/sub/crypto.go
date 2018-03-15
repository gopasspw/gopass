package sub

import (
	"context"
	"fmt"

	"github.com/justwatchcom/gopass/backend"
	gpgcli "github.com/justwatchcom/gopass/backend/crypto/gpg/cli"
	"github.com/justwatchcom/gopass/backend/crypto/gpg/openpgp"
	"github.com/justwatchcom/gopass/backend/crypto/plain"
	"github.com/justwatchcom/gopass/backend/crypto/xc"
	"github.com/justwatchcom/gopass/utils/agent/client"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
)

func (s *Store) initCryptoBackend(ctx context.Context) error {
	cb, err := GetCryptoBackend(ctx, s.url.Crypto, s.cfgdir, s.agent)
	if err != nil {
		return err
	}
	s.crypto = cb
	return nil
}

// GetCryptoBackend initialized the correct crypto backend
func GetCryptoBackend(ctx context.Context, cb backend.CryptoBackend, cfgdir string, agent *client.Client) (backend.Crypto, error) {
	switch cb {
	case backend.GPGCLI:
		out.Debug(ctx, "Using Crypto Backend: gpg-cli")
		return gpgcli.New(ctx, gpgcli.Config{
			Umask: fsutil.Umask(),
			Args:  gpgcli.GPGOpts(),
		})
	case backend.XC:
		out.Debug(ctx, "Using Crypto Backend: xc (EXPERIMENTAL)")
		return xc.New(cfgdir, agent)
	case backend.Plain:
		out.Debug(ctx, "Using Crypto Backend: plain (NO ENCRYPTION)")
		return plain.New(), nil
	case backend.OpenPGP:
		out.Debug(ctx, "Using Crypto Backend: openpgp (ALPHA)")
		return openpgp.New(ctx)
	default:
		return nil, fmt.Errorf("no valid crypto backend selected")
	}
}
