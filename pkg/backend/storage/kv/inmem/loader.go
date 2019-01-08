package inmem

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
)

const (
	name = "inmem"
)

func init() {
	backend.RegisterStorage(backend.InMem, name, &loader{})
}

type loader struct{}

// New implements backend.StorageLoader
func (l loader) New(ctx context.Context, url *backend.URL) (backend.Storage, error) {
	return New(), nil
}

func (l loader) String() string {
	return name
}
