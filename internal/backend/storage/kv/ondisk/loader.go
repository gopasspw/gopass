package ondisk

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
)

const (
	name = "ondisk"
)

func init() {
	backend.RegisterStorage(backend.OnDisk, name, &loader{})
	backend.RegisterRCS(backend.OnDiskRCS, name, &loader{})
}

type loader struct{}

// New creates a new ondisk loader
func (l loader) New(ctx context.Context, url *backend.URL) (backend.Storage, error) {
	be, err := New(url.Path)
	out.Debug(ctx, "Using Storage Backend: %s", be.String())
	return be, err
}

// Open loads an existing ondisk repo
func (l loader) Open(ctx context.Context, path string) (backend.RCS, error) {
	be, err := New(path)
	out.Debug(ctx, "Using RCS Backend: %s", be.String())
	return be, err
}

// Clone loads an existing ondisk repo
// WARNING: DOES NOT SUPPORT CLONE (yet)
func (l loader) Clone(ctx context.Context, repo, path string) (backend.RCS, error) {
	be, err := New(path)
	out.Debug(ctx, "Using RCS Backend: %s", be.String())
	return be, err
}

// Init creates a new ondisk repo
func (l loader) Init(ctx context.Context, path, username, email string) (backend.RCS, error) {
	be, err := New(path)
	out.Debug(ctx, "Using RCS Backend: %s", be.String())
	return be, err
}

// String returns ondisk
func (l loader) String() string {
	return name
}
