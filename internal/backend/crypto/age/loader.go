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
	if s.Exists(ctx, OldIDFile) || fsutil.IsNonEmptyFile(OldKeyring) {
		if err := migrate(ctx, s); err != nil {
			out.Errorf(ctx, "Failed to migrate age backend: %s", err)
		}
		out.OKf(ctx, "Migrated age backend to new format")

		return nil
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
