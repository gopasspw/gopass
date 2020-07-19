package action

import (
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// Convert converts a store to a different set of backends
func (s *Action) Convert(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)

	store := c.String("store")
	move := c.Bool("move")
	storage := backend.StorageBackendFromName(c.String("storage"))
	crypto := backend.CryptoBackendFromName(c.String("crypto"))

	return s.Store.Convert(ctx, store, crypto, storage, move)
}
