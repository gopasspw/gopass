package inmem

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
)

func init() {
	backend.RegisterStorage(backend.InMem, "inmem", &loader{})
}

type loader struct{}

// New implements backend.StorageLoader
func (l loader) New(ctx context.Context, url *backend.URL) (backend.Storage, error) {
	return New(), nil
}
