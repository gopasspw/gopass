package sub

import (
	"context"

	"github.com/gopasspw/gopass/pkg/agent/client"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/pkg/errors"
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
	ctx = client.WithClient(ctx, agent)
	ctx = ctxutil.WithConfigDir(ctx, cfgdir)
	crypto, err := backend.NewCrypto(ctx, cb)
	if err != nil {
		return nil, errors.Wrapf(err, "unknown crypto backend")
	}
	return crypto, nil
}
