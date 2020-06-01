package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const (
	name = "gitcli"
)

func init() {
	backend.RegisterRCS(backend.GitCLI, name, &loader{})
}

type loader struct{}

// Open implements backend.RCSLoader
func (l loader) Open(ctx context.Context, path string) (backend.RCS, error) {
	return Open(path)
}

// Clone implements backend.RCSLoader
func (l loader) Clone(ctx context.Context, repo, path string) (backend.RCS, error) {
	return Clone(ctx, repo, path)
}

// Init implements backend.RCSLoader
func (l loader) InitRCS(ctx context.Context, path string) (backend.RCS, error) {
	return Init(ctx, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

func (l loader) Handles(path string) error {
	if !fsutil.IsDir(filepath.Join(path, ".git")) {
		return fmt.Errorf("no .git")
	}
	return nil
}

func (l loader) Priority() int {
	return 1
}
func (l loader) String() string {
	return name
}
