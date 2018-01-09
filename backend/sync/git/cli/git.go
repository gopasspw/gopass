package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// Git is a cli based git backend
type Git struct {
	path string
}

// Open creates a new git cli based git backend
func Open(path, gpg string) (*Git, error) {
	if !fsutil.IsDir(filepath.Join(path, ".git")) {
		return nil, fmt.Errorf("git repo does not exist")
	}
	return &Git{
		path: path,
	}, nil
}

// Clone clones an existing git repo and returns a new cli based git backend
// configured for this clone repo
func Clone(ctx context.Context, repo, path string) (*Git, error) {
	g := &Git{
		path: filepath.Dir(path),
	}
	if err := g.Cmd(ctx, "Clone", "clone", repo, path); err != nil {
		return nil, err
	}
	g.path = path
	return g, nil
}

// Init initializes this store's git repo
func Init(ctx context.Context, path, userName, userEmail string) (*Git, error) {
	g := &Git{
		path: path,
	}
	// the git repo may be empty (i.e. no branches, cloned from a fresh remote)
	// or already initialized. Only run git init if the folder is completely empty
	if !g.IsInitialized() {
		if err := g.Cmd(ctx, "Init", "init"); err != nil {
			return nil, errors.Errorf("Failed to initialize git: %s", err)
		}
	}

	// initialize the local git config
	if err := g.InitConfig(ctx, userName, userEmail); err != nil {
		return g, errors.Errorf("failed to configure git: %s", err)
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
		return g, errors.Wrapf(err, "failed to commit changes to git")
	}

	return g, nil
}

// Cmd runs an git command
func (g *Git) Cmd(ctx context.Context, name string, args ...string) error {
	buf := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, "git", args[0:]...)
	cmd.Dir = g.path
	cmd.Stdout = buf
	cmd.Stderr = buf

	if ctxutil.IsDebug(ctx) || ctxutil.IsVerbose(ctx) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	out.Debug(ctx, "store.%s: %s %+v (%s)", name, cmd.Path, cmd.Args, g.path)

	if err := cmd.Run(); err != nil {
		out.Debug(ctx, "Output of '%s %+v': '%s'", cmd.Path, cmd.Args, buf.String())
		return errors.Wrapf(err, "failed to run command %s %+v", cmd.Path, cmd.Args)
	}

	return nil
}

// Name returns git
func (g *Git) Name() string {
	return "git"
}

// Version returns the git version as major, minor and patch level
func (g *Git) Version(ctx context.Context) semver.Version {
	v := semver.Version{}

	cmd := exec.CommandContext(ctx, "git", "version")
	cmdout, err := cmd.Output()
	if err != nil {
		out.Debug(ctx, "Failed to run 'git version': %s", err)
		return v
	}

	svStr := strings.TrimPrefix(string(cmdout), "git version ")
	if p := strings.Fields(svStr); len(p) > 0 {
		svStr = p[0]
	}

	sv, err := semver.ParseTolerant(svStr)
	if err != nil {
		out.Debug(ctx, "Failed to parse '%s' as semver: %s", svStr, err)
		return v
	}
	return sv
}

// IsInitialized returns true if this stores has an (probably) initialized .git folder
func (g *Git) IsInitialized() bool {
	return fsutil.IsFile(filepath.Join(g.path, ".git", "config"))
}

// Add adds the listed files to the git index
func (g *Git) Add(ctx context.Context, files ...string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	for i := range files {
		files[i] = strings.TrimPrefix(files[i], g.path+"/")
	}

	args := []string{"add", "--all"}
	args = append(args, files...)

	return g.Cmd(ctx, "gitAdd", args...)
}

// HasStagedChanges returns true if there are any staged changes which can be committed
func (g *Git) HasStagedChanges(ctx context.Context) bool {
	if err := g.Cmd(ctx, "gitDiffIndex", "diff-index", "--quiet", "HEAD"); err != nil {
		return true
	}
	return false
}

// Commit creates a new git commit with the given commit message
func (g *Git) Commit(ctx context.Context, msg string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	if !g.HasStagedChanges(ctx) {
		return store.ErrGitNothingToCommit
	}

	return g.Cmd(ctx, "gitCommit", "commit", "-m", msg)
}

// PushPull pushes the repo to it's origin.
// optional arguments: remote and branch
func (g *Git) PushPull(ctx context.Context, op, remote, branch string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "master"
	}

	if v, err := g.ConfigGet(ctx, "remote."+remote+".url"); err != nil || v == "" {
		return store.ErrGitNoRemote
	}

	if err := g.Cmd(ctx, "gitPush", "pull", remote, branch); err != nil {
		if op == "pull" {
			return err
		}
		out.Yellow(ctx, "Failed to pull before git push: %s", err)
	}
	if op == "pull" {
		return nil
	}

	return g.Cmd(ctx, "gitPush", "push", remote, branch)
}

// Push pushes to the git remote
func (g *Git) Push(ctx context.Context, remote, branch string) error {
	return g.PushPull(ctx, "push", remote, branch)
}

// Pull pulls from the git remote
func (g *Git) Pull(ctx context.Context, remote, branch string) error {
	return g.PushPull(ctx, "pull", remote, branch)
}

// AddRemote adds a new remote
func (g *Git) AddRemote(ctx context.Context, remote, url string) error {
	return g.Cmd(ctx, "gitAddRemote", "remote", "add", remote, url)
}
