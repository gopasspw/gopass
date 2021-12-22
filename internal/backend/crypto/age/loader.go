package age

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
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
	return New()
}

func (l loader) Handles(ctx context.Context, s backend.Storage) error {
	if s.Exists(ctx, OldIDFile) || s.Exists(ctx, OldKeyring) {
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
