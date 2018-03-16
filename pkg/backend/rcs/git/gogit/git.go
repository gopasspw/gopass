package gogit

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/pkg/errors"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	// Stdout is exported for mocking / redirecting
	Stdout io.Writer = os.Stdout
)

// Git is a go-git.v4 based git client
type Git struct {
	path string
	repo *git.Repository
	wt   *git.Worktree
}

// Open tries to open an existing git repo on disk
func Open(path string) (*Git, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	return &Git{
		path: path,
		repo: r,
		wt:   w,
	}, nil
}

// Clone tries to clone an git repo from a remote
func Clone(ctx context.Context, repo, path string) (*Git, error) {
	r, err := git.PlainCloneContext(ctx, path, false, &git.CloneOptions{
		URL:      repo,
		Progress: Stdout,
	})
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	return &Git{
		path: path,
		repo: r,
		wt:   w,
	}, nil
}

// Init creates a new git repo on disk
func Init(ctx context.Context, path string) (*Git, error) {
	g := &Git{
		path: path,
	}

	if !g.IsInitialized() {
		r, err := git.PlainInit(path, false)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize git: %s", err)
		}
		g.repo = r
		wt, err := g.repo.Worktree()
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get worktree: %s", err)
		}
		g.wt = wt
	}

	// add current content of the store
	if err := g.Add(ctx, g.path); err != nil {
		return g, errors.Wrapf(err, "failed to add '%s' to git", g.path)
	}

	// commit if there is something to commit
	if !g.HasStagedChanges(ctx) {
		out.Debug(ctx, "No staged changes")
		return g, nil
	}

	if err := g.Commit(ctx, "Add current content of password store"); err != nil {
		return nil, errors.Wrapf(err, "failed to commit changes to git")
	}

	return g, nil
}

// Version just returns the (static) go-git version
func (g *Git) Version(context.Context) semver.Version {
	return semver.Version{
		Major: 4,
		Build: []string{"go-git"},
	}
}

// IsInitialized returns true if this is a valid repo
func (g *Git) IsInitialized() bool {
	_, err := git.PlainOpen(g.path)
	return err == nil
}

// Add adds any number of files to git
func (g *Git) Add(ctx context.Context, files ...string) error {
	if len(files) < 1 || (len(files) == 1 && files[0] == g.path) {
		return g.addAll()
	}
	for _, file := range files {
		if strings.HasPrefix(file, g.path) {
			file = strings.TrimPrefix(file, g.path+string(filepath.Separator))
		}
		_, err := g.wt.Add(file)
		if err != nil {
			return errors.Wrapf(err, "failed to add file '%s': %s", file, err)
		}
	}
	return nil
}

func (g *Git) addAll() error {
	files := make([]string, 0, 10)
	err := filepath.Walk(g.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != g.path {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "." || info.Name() == ".." {
			return nil
		}
		files = append(files, strings.TrimPrefix(path, g.path+string(filepath.Separator)))
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "failed to walk '%s': %s", g.path, err)
	}

	for _, file := range files {
		_, err := g.wt.Add(file)
		if err != nil {
			return errors.Wrapf(err, "failed to add file '%s': %s", file, err)
		}
	}
	return nil
}

// HasStagedChanges retures true if there are changes which can be committed
func (g *Git) HasStagedChanges(ctx context.Context) bool {
	st, err := g.wt.Status()
	if err != nil {
		out.Red(ctx, "Error: Unable to get status: %s", err)
		return false
	}
	return !st.IsClean()
}

// Commit creates a new commit
func (g *Git) Commit(ctx context.Context, msg string) error {
	if !g.HasStagedChanges(ctx) {
		return store.ErrGitNothingToCommit
	}

	_, err := g.wt.Commit(msg, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  os.Getenv("GIT_AUTHOR_NAME"),
			Email: os.Getenv("GIT_AUTHOR_EMAIL"),
			When:  time.Now(),
		},
	})
	return err
}

// PushPull will first pull from the remote and then push any changes
func (g *Git) PushPull(ctx context.Context, op, remote, branch string) error {
	if err := g.Pull(ctx, remote, branch); err != nil {
		if op == "pull" {
			return err
		}
		out.Yellow(ctx, "Failed to pull before git push: %s", err)
	}

	if op == "pull" {
		return nil
	}
	return g.Push(ctx, remote, branch)
}

// Pull will pull any changes
func (g *Git) Pull(ctx context.Context, remote, branch string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "master"
	}

	cfg, err := g.repo.Config()
	if err != nil {
		return errors.Wrapf(err, "Failed to get git config: %s", err)
	}
	if _, found := cfg.Remotes[remote]; !found {
		return store.ErrGitNoRemote
	}

	return g.wt.PullContext(ctx, &git.PullOptions{
		RemoteName:    remote,
		ReferenceName: plumbing.ReferenceName(branch),
		Progress:      Stdout,
	})
}

// Push will push any changes to the remote
func (g *Git) Push(ctx context.Context, remote, branch string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	if remote == "" {
		remote = "origin"
	}

	cfg, err := g.repo.Config()
	if err != nil {
		return errors.Wrapf(err, "Failed to get git config: %s", err)
	}
	if _, found := cfg.Remotes[remote]; !found {
		return store.ErrGitNoRemote
	}
	return g.repo.PushContext(ctx, &git.PushOptions{
		RemoteName: remote,
		Progress:   Stdout,
	})
}

// Cmd is not supported and will go away eventually
func (g *Git) Cmd(context.Context, string, ...string) error {
	return fmt.Errorf("not supported")
}

// InitConfig is not yet implemented
func (g *Git) InitConfig(context.Context, string, string) error {
	return fmt.Errorf("not supported")
}

// AddRemote adds a new remote
func (g *Git) AddRemote(ctx context.Context, remote, url string) error {
	_, err := g.repo.CreateRemote(&config.RemoteConfig{
		Name: remote,
		URLs: []string{url},
	})
	return err
}

// Name returns go-git
func (g *Git) Name() string {
	return "go-git"
}

// Revisions is not implemented
func (g *Git) Revisions(context.Context, string) ([]backend.Revision, error) {
	return nil, fmt.Errorf("not yet implemented for %s", g.Name())
}

// GetRevision is not implemented
func (g *Git) GetRevision(context.Context, string, string) ([]byte, error) {
	return nil, fmt.Errorf("not yet implemented for %s", g.Name())
}
