// Package jjfs implements a jj cli based RCS backend.
package jjfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// JJFS is a cli based jj backend.
type JJFS struct {
	fs *fs.Store
}

// New creates a new jj cli based jj backend.
func New(path string) (*JJFS, error) {
	return &JJFS{
			fs: fs.New(path),
		},
		nil
}

// Init initializes this store's jj repo.
func Init(ctx context.Context, path, userName, userEmail string) (*JJFS, error) {
	j := &JJFS{
		fs: fs.New(path),
	}

	if !j.IsInitialized() {
		if err := j.Cmd(ctx, "Init", "init", "--git-repo", "."); err != nil {
			return nil, fmt.Errorf("failed to initialize jj: %w", err)
		}
		out.Printf(ctx, "jj initialized at %s", j.fs.Path())
	}

	if err := j.Add(ctx, j.fs.Path()); err != nil {
		return j, fmt.Errorf("failed to add %q to jj: %w", j.fs.Path(), err)
	}

	if err := j.Commit(ctx, "Add current content of password store"); err != nil {
		return j, fmt.Errorf("failed to commit changes to jj: %w", err)
	}

	return j, nil
}

func (j *JJFS) captureCmd(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	bufOut := &bytes.Buffer{}
	bufErr := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, "jj", args[0:]...)
	cmd.Dir = j.fs.Path()
	cmd.Stdout = bufOut
	cmd.Stderr = bufErr

	debug.Log("store.%s: %s %+v (%s)", name, cmd.Path, cmd.Args, j.fs.Path())
	err := cmd.Run()

	return bufOut.Bytes(), bufErr.Bytes(), err
}

// Cmd runs an jj command.
func (j *JJFS) Cmd(ctx context.Context, name string, args ...string) error {
	stdout, stderr, err := j.captureCmd(ctx, name, args...)
	if err != nil {
		debug.Log("CMD: %s %+v\nError: %s\nOutput:\n  Stdout: %q\n  Stderr: %q", name, args, err, string(stdout), string(stderr))

		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(stderr)))
	}

	return nil
}

// Name returns jj.
func (j *JJFS) Name() string {
	return "jjfs"
}

// Version returns the jj version.
func (j *JJFS) Version(ctx context.Context) semver.Version {
	v := semver.Version{}

	stdout, _, err := j.captureCmd(ctx, "version", "version")
	if err != nil {
		debug.Log("Failed to run 'jj version': %s", err)

		return v
	}

	sv, err := semver.ParseTolerant(string(stdout))
	if err != nil {
		debug.Log("Failed to parse %q as semver: %s", string(stdout), err)

		return v
	}

	return sv
}

// IsInitialized returns true if this stores has an (probably) initialized .jj folder.
func (j *JJFS) IsInitialized() bool {
	return fsutil.IsDir(j.fs.Path() + "/.jj")
}

// Add adds the listed files to the jj index.
func (j *JJFS) Add(ctx context.Context, files ...string) error {
	if !j.IsInitialized() {
		return store.ErrGitNotInit
	}

	for i := range files {
		files[i] = strings.TrimPrefix(files[i], j.fs.Path()+"/")
	}

	args := []string{"git", "add"}
	args = append(args, files...)

	return j.Cmd(ctx, "jjGitAdd", args...)
}

// TryAdd calls Add and returns nil if the git repo was not initialized.
func (j *JJFS) TryAdd(ctx context.Context, files ...string) error {
	err := j.Add(ctx, files...)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrGitNotInit) {
		debug.Log("JJFS not initialized. Ignoring.")

		return nil
	}

	return err
}

// Commit creates a new jj commit with the given commit message.
func (j *JJFS) Commit(ctx context.Context, msg string) error {
	if !j.IsInitialized() {
		return store.ErrGitNotInit
	}

	return j.Cmd(ctx, "jjCommit", "commit", "-m", msg)
}

// TryCommit calls commit and returns nil if there was nothing to commit or if the git repo was not initialized.
func (j *JJFS) TryCommit(ctx context.Context, msg string) error {
	err := j.Commit(ctx, msg)
	if err == nil {
		return nil
	}
	if errors.Is(err, store.ErrGitNothingToCommit) {
		debug.Log("Nothing to commit. Ignoring.")

		return nil
	}
	if errors.Is(err, store.ErrGitNotInit) {
		debug.Log("JJFS not initialized. Ignoring.")

		return nil
	}

	return err
}

// Push pushes to the git remote.
func (j *JJFS) Push(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}

	return j.Cmd(ctx, "jjGitPush", "git", "push", remote, branch)
}

// Pull pulls from the git remote.
func (j *JJFS) Pull(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")

		return nil
	}

	return j.Cmd(ctx, "jjGitPull", "git", "fetch", remote, branch)
}

// TryPush calls Push and returns nil if the git repo was not initialized.
func (j *JJFS) TryPush(ctx context.Context, remote, branch string) error {
	err := j.Push(ctx, remote, branch)
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, store.ErrGitNotInit):
		debug.Log("JJFS not initialized. Ignoring.")

		return nil
	case errors.Is(err, store.ErrGitNoRemote):
		debug.Log("JJFS has no remote. Ignoring.")

		return nil
	default:
		return err
	}
}

// Revisions will list all available revisions of the named entity.
func (j *JJFS) Revisions(ctx context.Context, name string) ([]backend.Revision, error) {
	args := []string{
		"log",
		"--revisions", "@",
		"--template",
		"commit_id \"\x1f\" author \"\x1f\" committer.timestamp() \"\x1f\" description \"\x1e\"",
		"--",
		name,
	}
	stdout, stderr, err := j.captureCmd(ctx, "Revisions", args...)
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
			if iv, err := strconv.ParseInt(p[2], 10, 64); err == nil {
				r.Date = time.Unix(iv, 0)
			}
		}

		if len(p) > 3 {
			r.Subject = p[3]
		}

		revs = append(revs, r)
	}

	debug.Log("Revisions for %s: %+v", name, revs)

	return revs, nil
}

// GetRevision will return the content of any revision of the named entity.
func (j *JJFS) GetRevision(ctx context.Context, name, revision string) ([]byte, error) {
	name = strings.TrimSpace(name)
	revision = strings.TrimSpace(revision)
	args := []string{
		"show",
		"--revision", revision,
		name,
	}
	stdout, stderr, err := j.captureCmd(ctx, "GetRevision", args...)
	if err != nil {
		debug.Log("Command failed: %s", string(stderr))

		return nil, err
	}

	return stdout, nil
}

// Status return the jj status output.
func (j *JJFS) Status(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := j.captureCmd(ctx, "jjStatus", "status")
	if err != nil {
		debug.Log("Command failed: %s\n%s", string(stdout), string(stderr))

		return nil, err
	}

	return stdout, nil
}

// Compact will run git gc.
func (j *JJFS) Compact(ctx context.Context) error {
	return j.Cmd(ctx, "jjGitGC", "git", "gc", "--aggressive")
}

// ListUntrackedFiles lists untracked files.
func (j *JJFS) ListUntrackedFiles(ctx context.Context) []string {
	stdout, _, err := j.captureCmd(ctx, "jjStatus", "status", "--no-patch")
	if err != nil {
		return []string{fmt.Sprintf("ERROR: %s", err)}
	}
	uf := []string{}
	for _, f := range strings.Split(string(stdout), "\n") {
		if f == "" || len(f) < 3 {
			continue
		}
		if f[0] == 'A' {
			uf = append(uf, strings.TrimSpace(f[2:]))
		}
	}

	return uf
}

// HasStagedChanges returns true if there are any staged changes which can be committed.
func (j *JJFS) HasStagedChanges(ctx context.Context) bool {
	stdout, _, err := j.captureCmd(ctx, "jjStatus", "status", "--no-patch")
	if err != nil {
		return false
	}

	return len(strings.TrimSpace(string(stdout))) > 0
}

// AddRemote adds a new remote.
func (j *JJFS) AddRemote(ctx context.Context, remote, url string) error {
	return j.Cmd(ctx, "jjGitRemoteAdd", "git", "remote", "add", remote, url)
}

// RemoveRemote removes a remote.
func (j *JJFS) RemoveRemote(ctx context.Context, remote string) error {
	return j.Cmd(ctx, "jjGitRemoteRemove", "git", "remote", "remove", remote)
}

// InitConfig initializes the git config.
func (j *JJFS) InitConfig(ctx context.Context, name, email string) error {
	return nil
}

// Get returns the content of a secret.
func (j *JJFS) Get(ctx context.Context, name string) ([]byte, error) {
	return j.fs.Get(ctx, name)
}

// Set writes the content of a secret.
func (j *JJFS) Set(ctx context.Context, name string, value []byte) error {
	return j.fs.Set(ctx, name, value)
}

// Delete removes a secret.
func (j *JJFS) Delete(ctx context.Context, name string) error {
	return j.fs.Delete(ctx, name)
}

// Exists checks if a secret exists.
func (j *JJFS) Exists(ctx context.Context, name string) bool {
	return j.fs.Exists(ctx, name)
}

// List returns a list of all secrets.
func (j *JJFS) List(ctx context.Context, prefix string) ([]string, error) {
	return j.fs.List(ctx, prefix)
}

// IsDir returns true if the given path is a directory.
func (j *JJFS) IsDir(ctx context.Context, name string) bool {
	return j.fs.IsDir(ctx, name)
}

// Prune removes a directory.
func (j *JJFS) Prune(ctx context.Context, prefix string) error {
	return j.fs.Prune(ctx, prefix)
}

// Link creates a symlink.
func (j *JJFS) Link(ctx context.Context, from, to string) error {
	return j.fs.Link(ctx, from, to)
}

// Path returns the path to the storage.
func (j *JJFS) Path() string {
	return j.fs.Path()
}

// Fsck checks the storage for errors.
func (j *JJFS) Fsck(ctx context.Context) error {
	return j.fs.Fsck(ctx)
}

// Move moves a file.
func (j *JJFS) Move(ctx context.Context, from, to string, del bool) error {
	return j.fs.Move(ctx, from, to, del)
}

func (j *JJFS) String() string {
	return j.fs.String()
}

// HasBranches returns true if the store has branches.
func (j *JJFS) HasBranches(ctx context.Context) bool {
	out, _, err := j.captureCmd(ctx, "HasBranches", "branch", "list")
	if err != nil {
		return false
	}
	return len(out) > 0
}
