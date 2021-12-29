package action

import (
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Convert converts a store to a different set of backends.
func (s *Action) Convert(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	store := c.String("store")
	move := c.Bool("move")
	storage, err := backend.StorageRegistry.Backend(c.String("storage"))
	if err != nil {
		return err
	}
	crypto, err := backend.CryptoRegistry.Backend(c.String("crypto"))
	if err != nil {
		return err
	}

	return s.Store.Convert(ctx, store, crypto, storage, move)
}
