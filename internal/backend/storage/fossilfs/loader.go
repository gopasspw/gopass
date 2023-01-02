package fossilfs

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const (
	name = "fossilfs"
)

func init() {
	backend.StorageRegistry.Register(backend.FossilFS, name, &loader{})
}

type loader struct{}

func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
	return New(path)
}

func (l loader) Open(ctx context.Context, path string) (backend.Storage, error) {
	return New(path)
}

func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	return Clone(ctx, repo, path)
}

func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	return Init(ctx, path, "", "")
}

func (l loader) Handles(ctx context.Context, path string) error {
	path = fsutil.ExpandHomedir(path)

	marker := filepath.Join(path, CheckoutMarker)
	if !fsutil.IsFile(marker) {
		return fmt.Errorf("no fossil checkout marker found at %s", marker)
	}

	return nil
}

func (l loader) Priority() int {
	return 2
}

func (l loader) String() string {
	return name
}
