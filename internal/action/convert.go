package action

import (
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Convert converts a store to a different set of backends.
func (s *Action) Convert(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	ctx = age.WithOnlyNative(ctx, true)

	store := c.String("store")
	move := c.Bool("move")

	sub, err := s.Store.GetSubStore(store)
	if err != nil {
		return fmt.Errorf("mount %q not found: %w", store, err)
	}

	oldStorage := sub.Storage().Name()

	storage, err := backend.StorageRegistry.Backend(oldStorage)
	if err != nil {
		return fmt.Errorf("unknown storage backend %q: %w", oldStorage, err)
	}
	if sv := c.String("storage"); sv != "" {
		var err error
		storage, err = backend.StorageRegistry.Backend(sv)
		if err != nil {
			return fmt.Errorf("unknown storage backend %q: %w", sv, err)
		}
	}

	oldCrypto := sub.Crypto().Name()

	crypto, err := backend.CryptoRegistry.Backend(oldCrypto)
	if err != nil {
		return fmt.Errorf("unknown crypto backend %q: %w", oldCrypto, err)
	}
	if sv := c.String("crypto"); sv != "" {
		var err error
		crypto, err = backend.CryptoRegistry.Backend(sv)
		if err != nil {
			return fmt.Errorf("unknown crypto backend %q: %w", sv, err)
		}
	}

	if oldCrypto == crypto.String() && oldStorage == storage.String() {
		out.Notice(ctx, "No conversion needed")

		return nil
	}

	if oldCrypto != crypto.String() {
		cbe, err := backend.NewCrypto(ctx, crypto)
		if err != nil {
			return err
		}

		if err := s.initCheckPrivateKeys(ctx, cbe); err != nil {
			return err
		}
		out.Printf(ctx, "Crypto %q has private keys", crypto.String())
	}

	out.Noticef(ctx, "Converting %q. Crypto: %q -> %q, Storage: %q -> %q", store, oldCrypto, crypto, oldStorage, storage)
	ok, err := termio.AskForBool(ctx, "Continue?", false)
	if err != nil {
		return err
	}
	if ctxutil.IsInteractive(ctx) && !ok {
		out.Notice(ctx, "Aborted")

		return nil
	}

	if err := s.Store.Convert(ctx, store, crypto, storage, move); err != nil {
		return fmt.Errorf("failed to convert %q: %w", store, err)
	}

	out.OKf(ctx, "Successfully converted %q", store)

	return nil
}
