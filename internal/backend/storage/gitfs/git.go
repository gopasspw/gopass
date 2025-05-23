// Package gitfs implements a git cli based RCS backend.
package gitfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gitconfig"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

type contextKey int

const (
	ctxKeyPathOverride contextKey = iota
)

func withPathOverride(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, ctxKeyPathOverride, path)
}

func getPathOverride(ctx context.Context, def string) string {
	if sv, ok := ctx.Value(ctxKeyPathOverride).(string); ok && sv != "" {
		return sv
	}

	return def
}

// Git is a cli based git backend.
type Git struct {
	fs  *fs.Store
	cfg *gitconfig.Configs
}

// New creates a new git cli based git backend.
func New(path string) (*Git, error) {
	path = fsutil.ExpandHomedir(path)

	gitDir := filepath.Join(path, ".git")
	if !fsutil.IsDir(gitDir) {
		return nil, fmt.Errorf("git repo does not exist at %s", gitDir)
	}

	return &Git{
		fs:  fs.New(path),
		cfg: gitconfig.New().LoadAll(gitDir),
	}, nil
}

// Clone clones an existing git repo and returns a new cli based git backend
// configured for this clone repo.
func Clone(ctx context.Context, repo, path, userName, userEmail string) (*Git, error) {
	g := &Git{
		fs:  fs.New(path),
		cfg: gitconfig.New(),
	}

	if err := g.Cmd(withPathOverride(ctx, filepath.Dir(path)), "Clone", "clone", repo, path); err != nil {
		return nil, err
	}

	g.cfg.LoadAll(filepath.Join(path, ".git"))

	// initialize the local git config.
	if err := g.InitConfig(ctx, userName, userEmail); err != nil {
		return g, fmt.Errorf("failed to configure git: %w", err)
	}
	out.Printf(ctx, "git configured at %s", g.fs.Path())

	return g, nil
}

// Init initializes this store's git repo.
func Init(ctx context.Context, path, userName, userEmail string) (*Git, error) {
	g := &Git{
		fs:  fs.New(path),
		cfg: gitconfig.New(),
	}

	// the git repo may be empty (i.e. no branches, cloned from a fresh remote)
	// or already initialized. Only run git init if the folder is completely empty.
	if !g.IsInitialized() {
		if err := g.Cmd(ctx, "Init", "init"); err != nil {
			return nil, fmt.Errorf("failed to initialize git: %w", err)
		}
		out.Printf(ctx, "git initialized at %s", g.fs.Path())
	}

	g.cfg.LoadAll(filepath.Join(path, ".git"))

	if !ctxutil.IsGitInit(ctx) {
		return g, nil
	}

	// initialize the local git config.
	if err := g.InitConfig(ctx, userName, userEmail); err != nil {
		return g, fmt.Errorf("failed to configure git: %w", err)
	}
	out.Printf(ctx, "git configured at %s", g.fs.Path())

	// add current content of the store.
	if err := g.Add(ctx, g.fs.Path()); err != nil {
		return g, fmt.Errorf("failed to add %q to git: %w", g.fs.Path(), err)
	}

	// commit if there is something to commit.
	if !g.HasStagedChanges(ctx) {
		debug.Log("No staged changes")

		return g, nil
	}

	if err := g.Commit(ctx, "Add current content of password store"); err != nil {
		return g, fmt.Errorf("failed to commit changes to git: %w", err)
	}

	return g, nil
}

func (g *Git) captureCmd(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	bufOut := &bytes.Buffer{}
	bufErr := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, "git", args[0:]...)
	cmd.Dir = getPathOverride(ctx, g.fs.Path())
	cmd.Stdout = bufOut
	cmd.Stderr = bufErr

	debug.Log("store.%s: %s %+v (%s)", name, cmd.Path, cmd.Args, g.fs.Path())
	err := cmd.Run()

	return bufOut.Bytes(), bufErr.Bytes(), err
}

// Cmd runs an git command.
func (g *Git) Cmd(ctx context.Context, name string, args ...string) error {
	stdout, stderr, err := g.captureCmd(ctx, name, args...)
	if err != nil {
		debug.Log("CMD: %s %+v\nError: %s\nOutput:\n  Stdout: %q\n  Stderr: %q", name, args, err, string(stdout), string(stderr))

		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(stderr)))
	}

	return nil
}

// Name returns git.
func (g *Git) Name() string {
	return name
}

// Version returns the git version as major, minor and patch level.
func (g *Git) Version(ctx context.Context) semver.Version {
	v := semver.Version{}

	cmd := exec.CommandContext(ctx, "git", "version")
	cmdout, err := cmd.Output()
	if err != nil {
		debug.Log("Failed to run 'git version': %s", err)

		return v
	}

	svStr := strings.TrimPrefix(string(cmdout), "git version ")
	if p := strings.Fields(svStr); len(p) > 0 {
		svStr = p[0]
	}

	sv, err := semver.ParseTolerant(svStr)
	if err != nil {
		debug.Log("Failed to parse %q as semver: %s", svStr, err)

		return v
	}

	return sv
}

var reLeadingNumber = regexp.MustCompile(`^\s*(\d+)\D*`)

func parseVersion(sv string) (semver.Version, error) {
	if sv, err := semver.ParseTolerant(sv); err == nil {
		return sv, nil
	}

	parts := strings.SplitN(sv, ".", 3)
	if len(parts) == 3 {
		// try to extract the number-only prefix from the patch level
		// e.g. "windwows.2" from "2.42.0.windows.2"
		matches := reLeadingNumber.FindStringSubmatch(parts[2])
		if len(matches) > 0 {
			parts[2] = matches[1]
			sv = strings.Join(parts, ".")
		}
	}

	return semver.ParseTolerant(sv)
}

// IsInitialized returns true if this stores has an (probably) initialized .git folder.
func (g *Git) IsInitialized() bool {
	return fsutil.IsFile(filepath.Join(g.fs.Path(), ".git", "config"))
}

// Add adds the listed files to the git index.
func (g *Git) Add(ctx context.Context, files ...string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	for i := range files {
		files[i] = strings.TrimPrefix(files[i], g.fs.Path()+"/")
	}

	args := []string{"add", "--all", "--force"}
	args = append(args, files...)

	return g.Cmd(ctx, "gitAdd", args...)
}

// TryAdd calls Add and returns nil if the git repo was not initialized.
func (g *Git) TryAdd(ctx context.Context, files ...string) error {
	err := g.Add(ctx, files...)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrGitNotInit) {
		debug.Log("Git not initialized. Ignoring.")

		return nil
	}

	return err
}

// HasStagedChanges returns true if there are any staged changes which can be committed.
func (g *Git) HasStagedChanges(ctx context.Context) bool {
	if err := g.Cmd(ctx, "gitDiffIndex", "diff-index", "--quiet", "HEAD"); err != nil {
		return true
	}

	return false
}

// ListUntrackedFiles lists untracked files.
func (g *Git) ListUntrackedFiles(ctx context.Context) []string {
	stdout, _, err := g.captureCmd(ctx, "gitLsFiles", "ls-files", ".", "--exclude-standard", "--others")
	if err != nil {
		return []string{fmt.Sprintf("ERROR: %s", err)}
	}
	uf := []string{}
	for _, f := range strings.Split(string(stdout), "\n") {
		if f == "" {
			continue
		}
		uf = append(uf, f)
	}

	return uf
}

// Commit creates a new git commit with the given commit message.
func (g *Git) Commit(ctx context.Context, msg string) error {
	if !g.IsInitialized() {
		return store.ErrGitNotInit
	}

	if !g.HasStagedChanges(ctx) {
		return store.ErrGitNothingToCommit
	}

	return g.Cmd(ctx, "gitCommit", "commit", fmt.Sprintf("--date=%d +00:00", ctxutil.GetCommitTimestamp(ctx).UTC().Unix()), "-m", msg)
}

// TryCommit calls commit and returns nil if there was nothing to commit or if the git repo was not initialized.
func (g *Git) TryCommit(ctx context.Context, msg string) error {
	err := g.Commit(ctx, msg)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrGitNothingToCommit) {
		debug.Log("Nothing to commit. Ignoring.")

		return nil
	}
	if errors.Is(err, store.ErrGitNotInit) {
		debug.Log("Git not initialized. Ignoring.")

		return nil
	}

	return err
}

func (g *Git) defaultRemote(ctx context.Context, branch string) string {
	opts, err := g.ConfigList(ctx)
	if err != nil {
		return "origin"
	}

	remote := opts["branch."+branch+".remote"]
	if remote == "" {
		return "origin"
	}

	needle := "remote." + remote + ".url"
	for k := range opts {
		if k == needle {
			return remote
		}
	}

	return "origin"
}

func (g *Git) defaultBranch(ctx context.Context) string {
	out, _, err := g.captureCmd(ctx, "defaultBranch", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil || string(out) == "" {
		// see https://github.com/github/renaming.
		return "main"
	}

	return strings.TrimSpace(string(out))
}

// PushPull pushes the repo to it's origin.
// optional arguments: remote and branch.
func (g *Git) PushPull(ctx context.Context, op, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}
	if !g.IsInitialized() {
		debug.Log("Git in %s is not initialized. Can not push/pull", g.Path())

		return store.ErrGitNotInit
	}

	if branch == "" {
		branch = g.defaultBranch(ctx)
	}

	if remote == "" {
		remote = g.defaultRemote(ctx, branch)
	}

	urlKey := "remote." + remote + ".url"
	if v, err := g.ConfigGet(ctx, urlKey); err != nil || v == "" {
		debug.Log("No value for %q found in config. Keys: %+v", urlKey, g.cfg.Keys())

		return store.ErrGitNoRemote
	}

	if err := g.Cmd(ctx, "gitPush", "pull", remote, branch); err != nil {
		if op == "pull" {
			return err
		}
		out.Warningf(ctx, "Failed to pull before git push: %s", err)
	}

	if op == "pull" {
		return nil
	}

	if uf := g.ListUntrackedFiles(ctx); len(uf) > 0 {
		out.Warningf(ctx, "Found untracked files: %+v", uf)
	}

	return g.Cmd(ctx, "gitPush", "push", remote, branch)
}

// TryPush calls Push and returns nil if the git repo was not initialized.
func (g *Git) TryPush(ctx context.Context, remote, branch string) error {
	err := g.Push(ctx, remote, branch)
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, store.ErrGitNotInit):
		debug.Log("Git not initialized. Ignoring.")

		return nil
	case errors.Is(err, store.ErrGitNoRemote):
		debug.Log("Git has no remote. Ignoring.")

		return nil
	default:
		return err
	}
}

// Push pushes to the git remote.
func (g *Git) Push(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}

	return g.PushPull(ctx, "push", remote, branch)
}

// Pull pulls from the git remote.
func (g *Git) Pull(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}

	return g.PushPull(ctx, "pull", remote, branch)
}

// AddRemote adds a new remote.
func (g *Git) AddRemote(ctx context.Context, remote, url string) error {
	return g.Cmd(ctx, "gitAddRemote", "remote", "add", remote, url)
}

// RemoveRemote removes a remote.
func (g *Git) RemoveRemote(ctx context.Context, remote string) error {
	return g.Cmd(ctx, "gitRemoveRemote", "remote", "remove", remote)
}

// Revisions will list all available revisions of the named entity
// see http://blog.lost-theory.org/post/how-to-parse-git-log-output/
// and https://git-scm.com/docs/git-log#_pretty_formats.
func (g *Git) Revisions(ctx context.Context, name string) ([]backend.Revision, error) {
	args := []string{
		"log",
		`--format=%H%x1f%an%x1f%ae%x1f%at%x1f%s%x1f%b%x1e`,
		"--",
		name,
	}
	stdout, stderr, err := g.captureCmd(ctx, "Revisions", args...)
	if err != nil {
		debug.Log("Command failed: %s", string(stderr))

		return nil, err
	}

	so := string(stdout)
	revs := make([]backend.Revision, 0, strings.Count(so, "\x1e"))
	for _, rev := range strings.Split(so, "\x1e") {
		rev = strings.TrimSpace(rev)
		if rev == "" {
			continue
		}

		p := strings.Split(rev, "\x1f")
		if len(p) < 1 {
			continue
		}

		r := backend.Revision{}
		r.Hash = p[0]
		if len(p) > 1 {
			r.AuthorName = p[1]
		}

		if len(p) > 2 {
			r.AuthorEmail = p[2]
		}

		if len(p) > 3 {
			if iv, err := strconv.ParseInt(p[3], 10, 64); err == nil {
				r.Date = time.Unix(iv, 0)
			}
		}

		if len(p) > 4 {
			r.Subject = p[4]
		}

		if len(p) > 5 {
			r.Body = p[5]
		}

		revs = append(revs, r)
	}

	debug.Log("Revisions for %s: %+v", name, revs)

	return revs, nil
}

// GetRevision will return the content of any revision of the named entity
// see https://git-scm.com/docs/git-log#_pretty_formats.
func (g *Git) GetRevision(ctx context.Context, name, revision string) ([]byte, error) {
	name = strings.TrimSpace(name)
	revision = strings.TrimSpace(revision)
	args := []string{
		"show",
		revision + ":" + name,
	}
	stdout, stderr, err := g.captureCmd(ctx, "GetRevision", args...)
	if err != nil {
		debug.Log("Command failed: %s", string(stderr))

		return nil, err
	}

	return stdout, nil
}

// Status return the git status output.
func (g *Git) Status(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := g.captureCmd(ctx, "GitStatus", "status")
	if err != nil {
		debug.Log("Command failed: %s\n%s", string(stdout), string(stderr))

		return nil, err
	}

	return stdout, nil
}

// Compact will run git gc.
func (g *Git) Compact(ctx context.Context) error {
	return g.Cmd(ctx, "gitGC", "gc", "--aggressive")
}
