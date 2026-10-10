package gitfs

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

const (
	name = "gitfs"
)

// ErrGitBinaryNotFound is returned when the git binary is not available on
// PATH. It signals that callers should fall back to the pure-Go go-git backend.
var ErrGitBinaryNotFound = errors.New("git binary not found in PATH")

func init() {
	backend.StorageRegistry.Register(backend.GitFS, name, &loader{})
}

type loader struct{}

func (l loader) New(ctx context.Context, path string) (backend.Storage, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, ErrGitBinaryNotFound
	}

	return New(path)
}

// Open implements backend.RCSLoader.
func (l loader) Open(ctx context.Context, path string) (backend.Storage, error) {
	return New(path)
}

// Clone implements backend.RCSLoader.
func (l loader) Clone(ctx context.Context, repo, path string) (backend.Storage, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, ErrGitBinaryNotFound
	}

	return Clone(ctx, repo, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

// Init implements backend.RCSLoader.
func (l loader) Init(ctx context.Context, path string) (backend.Storage, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, ErrGitBinaryNotFound
	}

	return Init(ctx, path, termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
}

func (l loader) Handles(ctx context.Context, path string) error {
	path = fsutil.ExpandHomedir(path)
	if !fsutil.IsDir(filepath.Join(path, ".git")) {
		return fmt.Errorf("no .git at %s", path)
	}

	if _, err := exec.LookPath("git"); err != nil {
		return ErrGitBinaryNotFound
	}

	return nil
}

func (l loader) Priority() int {
	return 11
}

func (l loader) String() string {
	return name
}
