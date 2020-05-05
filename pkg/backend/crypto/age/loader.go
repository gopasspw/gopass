package age

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
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

func (l loader) String() string {
	return name
}
