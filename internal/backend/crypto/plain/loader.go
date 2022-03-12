package plain

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	name = "plain"
)

func init() {
	backend.CryptoRegistry.Register(backend.Plain, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	debug.Log("Using Crypto Backend: %s (NO ENCRYPTION)", name)

	return New(), nil
}

func (l loader) Handles(ctx context.Context, s backend.Storage) error {
	if s.Exists(ctx, IDFile) {
		return nil
	}

	return fmt.Errorf("not supported")
}

func (l loader) Priority() int {
	return 1000
}

func (l loader) String() string {
	return name
}
