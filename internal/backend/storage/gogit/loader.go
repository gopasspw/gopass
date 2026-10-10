// Package gogit implements a pure-Go git RCS backend based on
// github.com/go-git/go-git/v5. It is an opt-in alternative to the git CLI
// backend (gitfs) and is selected automatically when the git binary is not
// available on PATH.
//
// Unlike gitfs this backend does not need the git binary at runtime. See
// docs/backends/gogit.md for the feature parity matrix, the conflict policy
// and the known limitations (commit signing, credential helpers, git hooks).
package gogit

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

const (
	name = "gogit"
)

func init() {
	backend.StorageRegistry.Register(backend.GoGit, name, &loader{})
}

type loader struct{}

func (l loader) String() string {
	return name
}

// Priority must be higher (later) than gitfs (11) so that, when both the git
// binary and go-git are available, auto-detection still prefers the CLI
// backend. Auto-fallback to go-git is driven by the gitfs loader failing with
// ErrGitBinaryNotFound when git is absent, not by priority alone.
func (l loader) Priority() int {
	return 12
}

func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
	return New(path)
}

func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	return Init(ctx, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	return Clone(ctx, repo, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

// Handles returns nil iff path contains a .git directory. Unlike gitfs it does
// NOT require a git binary.
func (l loader) Handles(_ context.Context, path string) error {
	path = fsutil.ExpandHomedir(path)
	if !fsutil.IsDir(filepath.Join(path, ".git")) {
		return fmt.Errorf("no .git at %s", path)
	}

	return nil
}
