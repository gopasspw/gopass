package gitfs

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

const (
	name = "gitfs"
)

func init() {
	backend.StorageRegistry.Register(backend.GitFS, name, &loader{})
}

type loader struct{}

func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
	return New(path)
}

// Open implements backend.RCSLoader.
func (l loader) Open(ctx context.Context, path string) (backend.Storage, error) {
	return New(path)
}

// Clone implements backend.RCSLoader.
func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	return Clone(ctx, repo, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

// Init implements backend.RCSLoader.
func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	return Init(ctx, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

func (l loader) Handles(ctx context.Context, path string) error {
	path = fsutil.ExpandHomedir(path)
	gitPath := filepath.Join(path, ".git")
	if !fsutil.IsDir(gitPath) && !fsutil.IsFile(gitPath) {
		return fmt.Errorf("no .git at %s", path)
	}

	return nil
}

func (l loader) Priority() int {
	return 1
}

func (l loader) String() string {
	return name
}
