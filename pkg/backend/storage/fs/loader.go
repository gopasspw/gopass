package fs

import (
	"context"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/out"
)

const (
	name = "fs"
)

func init() {
	backend.RegisterStorage(backend.FS, name, &loader{})
}

type loader struct{}

// New implements backend.StorageLoader
func (l loader) New(ctx context.Context, url *backend.URL) (backend.Storage, error) {
	be := New(url.Path)
	out.Debug(ctx, "Using Storage Backend: %s", be.String())
	return be, nil
}

func (l loader) String() string {
	return name
}
