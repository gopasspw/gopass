package gogit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// GoGit is a pure-Go git backend. File operations delegate to the fs backend,
// RCS operations use go-git.
type GoGit struct {
	fs   *fs.Store
	repo *git.Repository
}

var _ backend.Storage = (*GoGit)(nil)

// New opens an existing git repository at path.
func New(path string) (*GoGit, error) {
	path = fsutil.ExpandHomedir(path)

	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: false})
	if err != nil {
		return nil, fmt.Errorf("no git repository at %s: %w", path, err)
	}

	return &GoGit{fs: fs.New(path), repo: repo}, nil
}

// Clone clones an existing git repo and returns a new GoGit backend.
func Clone(ctx context.Context, repo, path, userName, userEmail string) (*GoGit, error) {
	path = fsutil.ExpandHomedir(path)
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create %q: %w", path, err)
	}

	r, err := git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		URL:      repo,
		Auth:     authForURL(ctx, repo),
		Progress: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone %q to %q: %w", repo, path, err)
	}

	g := &GoGit{fs: fs.New(path), repo: r}

	if err := g.InitConfig(ctx, userName, userEmail); err != nil {
		return g, fmt.Errorf("failed to configure git: %w", err)
	}
	out.Printf(ctx, "git (go-git) configured at %s", g.fs.Path())

	return g, nil
}

// Init initializes this store's git repo.
func Init(ctx context.Context, path, userName, userEmail string) (*GoGit, error) {
	path = fsutil.ExpandHomedir(path)
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create %q: %w", path, err)
	}

	var repo *git.Repository

	if fsutil.IsDir(filepath.Join(path, ".git")) {
		r, err := git.PlainOpen(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open git repo at %s: %w", path, err)
		}
		repo = r
	} else {
		r, err := git.PlainInitWithOptions(path, &git.PlainInitOptions{
			InitOptions: git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName("master")},
			Bare:        false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to init git repo: %w", err)
		}
		repo = r
		out.Printf(ctx, "git (go-git) initialized at %s", path)
	}

	g := &GoGit{fs: fs.New(path), repo: repo}

	if !ctxutil.IsGitInit(ctx) {
		return g, nil
	}

	if err := g.InitConfig(ctx, userName, userEmail); err != nil {
		return g, fmt.Errorf("failed to configure git: %w", err)
	}
	out.Printf(ctx, "git (go-git) configured at %s", g.fs.Path())

	// stage the current content of the store.
	if err := g.Add(ctx, g.fs.Path()); err != nil {
		return g, fmt.Errorf("failed to add %q to git: %w", g.fs.Path(), err)
	}

	if !g.HasStagedChanges(ctx) {
		debug.Log("No staged changes")

		return g, nil
	}

	if ctxutil.HasSetupRemote(ctx) {
		debug.Log("Skipping auto-commit during setup with specified remote")

		return g, nil
	}

	if err := g.Commit(ctx, "Add current content of password store"); err != nil && !errors.Is(err, store.ErrGitNothingToCommit) {
		return g, fmt.Errorf("failed to commit changes to git: %w", err)
	}

	return g, nil
}

// Name returns gogit.
func (g *GoGit) Name() string {
	return name
}

// Version returns the go-git module version.
func (g *GoGit) Version(_ context.Context) semver.Version {
	return debug.ModuleVersion("github.com/go-git/go-git/v5")
}

// Path returns the path to this storage.
func (g *GoGit) Path() string {
	return g.fs.Path()
}

// String implements fmt.Stringer.
func (g *GoGit) String() string {
	return fmt.Sprintf("gogit(%s,path:%s)", g.Version(context.Background()).String(), g.fs.Path())
}

// IsInitialized returns true if this store has a (probably) initialized .git folder.
func (g *GoGit) IsInitialized() bool {
	return fsutil.IsFile(filepath.Join(g.fs.Path(), ".git", "config"))
}
