// Package jjfs implements a jj cli based RCS backend.
package jjfs

import (
	"context"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

func init() {
	backend.StorageRegistry.Register(backend.JJFS, "jj", &loader{})
}

type loader struct{}

func (l loader) String() string {
	return "jjfs"
}

func (l loader) Priority() int {
	return 10
}

func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
	return New(path)
}

func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	return Init(ctx, path, "", "")
}

func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	return nil, backend.ErrNotSupported
}

func (l loader) Handles(ctx context.Context, path string) error {
	if fsutil.IsDir(path + "/.jj") {
		return nil
	}

	return backend.ErrNotSupported
}
