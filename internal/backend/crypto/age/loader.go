package age

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const (
	name = "age"
)

func init() {
	backend.CryptoRegistry.Register(backend.Age, name, &loader{})
}

type loader struct{}

func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	debug.Log("Using Crypto Backend: %s", name)

	return New(ctx)
}

func (l loader) Handles(ctx context.Context, s backend.Storage) error {
	// OldKeyring is meant to be in the config folder, not necessarily in the store
	oldKeyring := OldKeyringPath()
	if s.Exists(ctx, OldIDFile) || fsutil.IsNonEmptyFile(oldKeyring) {
		debug.Log("Starting migration of age backend. Found ID File at %s = %t. Migrating to %s. Found Keyring at %s = %t", OldIDFile, s.Exists(ctx, OldIDFile), IDFile, oldKeyring, fsutil.IsNonEmptyFile(oldKeyring))

		if err := migrate(ctx, s); err != nil {
			out.Errorf(ctx, "Failed to migrate age backend: %s", err)

			return err
		}
		out.OKf(ctx, "Migrated age backend to new format")
	}
	if s.Exists(ctx, IDFile) {
		return nil
	}

	return fmt.Errorf("not supported")
}

func (l loader) Priority() int {
	return 10
}

func (l loader) String() string {
	return name
}
