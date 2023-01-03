package fs

import (
	"context"
	"fmt"
	"os"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const (
	name = "fs"
)

func init() {
	backend.StorageRegistry.Register(backend.FS, name, &loader{})
}

type loader struct{}

// New implements backend.StorageLoader.
func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, err
	}
	be := New(path)
	debug.Log("Using Storage Backend: %s", be.String())

	return be, nil
}

func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, err
	}

	return l.New(ctx, path)
}

// Clone is a no-op.
func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	return l.New(ctx, path)
}

// Handles returns true if the given path is supported by this backend. Will always return
// true if the directory exists.
func (l loader) Handles(ctx context.Context, path string) error {
	path = fsutil.ExpandHomedir(path)

	if fsutil.IsDir(path) {
		return nil
	}

	return fmt.Errorf("dir not found")
}

// Priority returns the priority of this backend. Should always be higher than
// the more specific ones, e.g. gitfs.
func (l loader) Priority() int {
	return 50
}

func (l loader) String() string {
	return name
}
