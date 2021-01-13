package ondisk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const (
	name = "ondisk"
)

func init() {
	backend.RegisterStorage(backend.OnDisk, name, &loader{})
}

type loader struct{}

// New creates a new ondisk loader
func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
	be, err := New(ctx, path)
	debug.Log("Using Storage Backend %p: %s", be, path)
	return be, err
}

// Open loads an existing ondisk repo
func (l loader) Open(ctx context.Context, path string) (backend.Storage, error) {
	be, err := New(ctx, path)
	debug.Log("Using RCS Backend: %s", be.String())
	return be, err
}

// Clone loads an existing ondisk repo
func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	be, err := New(ctx, path)
	debug.Log("Using RCS Backend %p: %s", be, be.String())
	if err := be.SetRemote(ctx, repo); err != nil {
		return nil, err
	}
	if err := be.Pull(ctx, "", ""); err != nil {
		return nil, err
	}
	return be, err
}

func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}
	be, err := New(ctx, path)
	debug.Log("Using RCS Backend %p: %s", be, be.String())
	return be, err
}

func (l loader) Handles(path string) error {
	if fsutil.IsFile(filepath.Join(path, idxFile)) {
		return nil
	}
	return fmt.Errorf("not supported")
}

func (l loader) Priority() int {
	return 49
}

// String returns ondisk
func (l loader) String() string {
	return name
}
