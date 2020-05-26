package age

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
)

const (
	name = "age"
)

func init() {
	backend.RegisterCrypto(backend.Age, name, &loader{})
}

type loader struct{}

func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	out.Debug(ctx, "Using Crypto Backend: %s", name)
	return New()
}

func (l loader) Handles(s backend.Storage) error {
	if s.Exists(context.TODO(), IDFile) {
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
