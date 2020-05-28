package leaf

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
)

func (s *Store) initCryptoBackend(ctx context.Context) error {
	cb, err := backend.DetectCrypto(ctx, s.storage)
	if err != nil {
		return err
	}
	s.crypto = cb
	return nil
}
