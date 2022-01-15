package fossilfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

const (
	// CheckoutMarker is the marker file that indicates a fossil checkout.
	CheckoutMarker = ".fslckout"
)

// Fossil is a storage backend for Fossil.
type Fossil struct {
	fs *fs.Store
}

// New instantiates a new Fossil store.
func New(path string) (*Fossil, error) {
	marker := filepath.Join(path, CheckoutMarker)
	if !fsutil.IsFile(marker) {
		return nil, fmt.Errorf("no fossil checkout marker found at %s", marker)
	}
	return &Fossil{
		fs: fs.New(path),
	}, nil
}

// Clone opens a new fossil checkout.
func Clone(ctx context.Context, repo, path string) (*Fossil, error) {
	f := &Fossil{
		fs: fs.New(path),
	}
	// we use open instead of clone, since that automatically clones, if necessary
	args := []string{
		"open", repo,
		"--workdir", path,
	}
	// the --repodir option only makes sense if the REPOSITORY argument is a URI that begins with http:, https:, ssh:, or file:
	if strings.HasPrefix(repo, "http:") || strings.HasPrefix(repo, "https:") || strings.HasPrefix(repo, "ssh:") || strings.HasPrefix(repo, "file:") {
		args = append(args, "--repodir", filepath.Dir(path))
	}
	if err := f.Cmd(withPathOverride(ctx, filepath.Dir(path)), "Clone", args...); err != nil {
		return nil, err
	}

	// initialize the local fossil config.
	if err := f.InitConfig(ctx, "", ""); err != nil {
		return f, fmt.Errorf("failed to configure git: %s", err)
	}
	out.Printf(ctx, "fossil configured at %s", f.fs.Path())

	return f, nil
}

// Init initializes this store's fossil repo.
func Init(ctx context.Context, path, _, _ string) (*Fossil, error) {
	f := &Fossil{
		fs: fs.New(path),
	}
	// the fossil repo may be empty (i.e. no branches, cloned from a fresh remote)
	// or already initialized. Only run fossil init if the folder is completely empty.
	if !f.IsInitialized() {
		repo := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".fossil")
		if err := f.Cmd(ctx, "Init", "init", repo); err != nil {
			return nil, fmt.Errorf("failed to initialize fossil in %s: %s", repo, err)
		}
		if err := f.Cmd(ctx, "Open", "open", repo); err != nil {
			return nil, fmt.Errorf("failed to open fossil in %s: %s", repo, err)
		}
		out.Printf(ctx, "fossil initialized at %s", f.fs.Path())
	}

	// TODO rename to IsRCSInitialized
	if !ctxutil.IsGitInit(ctx) {
		return f, nil
	}

	// initialize the local fossil config.
	if err := f.InitConfig(ctx, "", ""); err != nil {
		return f, fmt.Errorf("failed to configure fossil: %s", err)
	}
	out.Printf(ctx, "fossil configured at %s", f.fs.Path())

	// add current content of the store.
	if err := f.Add(ctx, f.fs.Path()); err != nil {
		return f, fmt.Errorf("failed to add %q to fossil: %w", f.fs.Path(), err)
	}

	// commit if there is something to commit.
	if !f.HasStagedChanges(ctx) {
		debug.Log("No staged changes")
		return f, nil
	}

	if err := f.Commit(ctx, "Add current content of password store"); err != nil {
		return f, fmt.Errorf("failed to commit changes to fossil: %w", err)
	}

	return f, nil
}

func (f *Fossil) captureCmd(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	bufOut := &bytes.Buffer{}
	bufErr := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, "fossil", args[0:]...)
	cmd.Dir = getPathOverride(ctx, f.fs.Path())
	cmd.Stdout = bufOut
	cmd.Stderr = bufErr

	if debug.IsEnabled() && ctxutil.IsVerbose(ctx) {
		cmd.Stdout = io.MultiWriter(bufOut, os.Stdout)
		cmd.Stderr = io.MultiWriter(bufErr, os.Stderr)
	}

	debug.Log("fossil.%s: %s %+v (%s)", name, cmd.Path, cmd.Args, f.fs.Path())
	err := cmd.Run()
	return bufOut.Bytes(), bufErr.Bytes(), err
}

// Cmd runs an fossil command.
func (f *Fossil) Cmd(ctx context.Context, name string, args ...string) error {
	stdout, stderr, err := f.captureCmd(ctx, name, args...)
	if err != nil {
		debug.Log("CMD: %s %+v\nError: %s\nOutput:\n  Stdout: %q\n  Stderr: %q", name, args, err, string(stdout), string(stderr))
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(stderr)))
	}

	return nil
}

// Name returns 'fossil'.
func (f *Fossil) Name() string {
	return name
}

// Version returns the fossil version as major, minor and patch level.
func (f *Fossil) Version(ctx context.Context) semver.Version {
	v := semver.Version{}

	cmd := exec.CommandContext(ctx, "fossil", "version")
	cmdout, err := cmd.Output()
	if err != nil {
		debug.Log("Failed to run 'fossil version': %s", err)
		return v
	}

	svStr := strings.TrimPrefix(string(cmdout), "This is fossil version ")
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

// IsInitialized returns true if this stores has an (probably) initialized Fossil chkecout.
func (f *Fossil) IsInitialized() bool {
	return fsutil.IsFile(filepath.Join(f.fs.Path(), CheckoutMarker))
}

// Add adds the listed files to the fossil index.
func (f *Fossil) Add(ctx context.Context, files ...string) error {
	if !f.IsInitialized() {
		// TODO should rename to ErrRCSNotInitialized
		return store.ErrGitNotInit
	}

	for i := range files {
		files[i] = strings.TrimPrefix(files[i], f.fs.Path()+"/")
	}

	args := []string{"add", "--force", "--dotfiles"}
	args = append(args, files...)

	return f.Cmd(ctx, "fossilAdd", args...)
}

// HasStagedChanges returns true if there are any staged changes which can be committed.
func (f *Fossil) HasStagedChanges(ctx context.Context) bool {
	s, err := f.getStatus(ctx)
	if err != nil {
		// TODO should return an error
		return true
	}
	return s.Staged().Len() > 0
}

// ListUntrackedFiles lists untracked files.
func (f *Fossil) ListUntrackedFiles(ctx context.Context) []string {
	s, err := f.getStatus(ctx)
	if err != nil {
		// TODO should return an error
		return []string{fmt.Sprintf("ERROR: %s", err)}
	}
	return s.Untracked().Elements()
}

// Commit creates a new fossil commit with the given commit message.
func (f *Fossil) Commit(ctx context.Context, msg string) error {
	if !f.IsInitialized() {
		return store.ErrGitNotInit
	}

	if !f.HasStagedChanges(ctx) {
		return store.ErrGitNothingToCommit
	}

	return f.Cmd(
		ctx,
		"fossilCommit",
		"commit",
		"--date-override",
		ctxutil.GetCommitTimestamp(ctx).UTC().Format("2006-01-02T15:04:05.000"),
		"--no-warnings",
		"-m",
		msg,
	)
}

// PushPull pushes the repo to it's origin.
// optional arguments: remote and branch.
func (f *Fossil) PushPull(ctx context.Context, op, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")
		return nil
	}
	if !f.IsInitialized() {
		return store.ErrGitNotInit
	}

	if uf := f.ListUntrackedFiles(ctx); len(uf) > 0 {
		out.Warningf(ctx, "Found untracked files: %+v", uf)
	}
	return f.Cmd(ctx, "fossilUpdate", "update")
}

// Push pushes to the fossil remote.
func (f *Fossil) Push(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")
		return nil
	}
	return f.PushPull(ctx, "push", remote, branch)
}

// Pull pulls from the fossil remote.
func (f *Fossil) Pull(ctx context.Context, remote, branch string) error {
	if ctxutil.IsNoNetwork(ctx) {
		debug.Log("Skipping network ops. NoNetwork=true")
		return nil
	}
	return f.PushPull(ctx, "pull", remote, branch)
}

// AddRemote adds a new remote.
func (f *Fossil) AddRemote(ctx context.Context, remote, url string) error {
	return f.Cmd(ctx, "fossilAddRemote", "remote", "add", remote, url)
}

// RemoveRemote removes a remote.
func (f *Fossil) RemoveRemote(ctx context.Context, remote string) error {
	return f.Cmd(ctx, "fossilRemoveRemote", "remote", "delete", remote)
}

// Revisions will list all available revisions of the named entity.
func (f *Fossil) Revisions(ctx context.Context, name string) ([]backend.Revision, error) {
	args := []string{
		"finfo",
		"-W",
		"0",
		name,
	}
	stdout, stderr, err := f.captureCmd(ctx, "Revisions", args...)
	if err != nil {
		debug.Log("Command failed: %s", string(stderr))
		return nil, err
	}

	revs := make([]backend.Revision, 0, strings.Count(string(stdout), "\n"))
	for _, line := range strings.Split(string(stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			debug.Log("empty line")
			continue
		}

		debug.Log("Parsing line: %s", line)
		body := line // retain full line for the body
		date, line, found := strings.Cut(line, " ")
		if !found {
			debug.Log("Failed to parse date")
			continue
		}
		rev, line, found := strings.Cut(line, " ")
		if !found {
			debug.Log("Failed to parse revision")
			continue
		}
		rev = strings.Trim(rev, "[]")
		subject, line, found := strings.Cut(line, "(")
		if !found {
			debug.Log("Failed to parse subject")
			continue
		}

		author, _, found := strings.Cut(line, ",")
		if !found {
			debug.Log("Failed to parse author")
			continue
		}
		author = strings.TrimPrefix(author, "user: ")

		ts, err := time.Parse("2006-01-02", date)
		if err != nil {
			debug.Log("Failed to parse date %s: %s", date, err)
			continue
		}

		r := backend.Revision{
			Hash:       rev,
			Date:       ts,
			Body:       body,
			Subject:    subject,
			AuthorName: author,
		}
		revs = append(revs, r)
	}
	return revs, nil
}

// GetRevision will return the content of any revision of the named entity.
func (f *Fossil) GetRevision(ctx context.Context, name, revision string) ([]byte, error) {
	name = strings.TrimSpace(name)
	revision = strings.TrimSpace(revision)
	args := []string{
		"cat",
		"-r",
		revision,
		name,
	}
	stdout, stderr, err := f.captureCmd(ctx, "GetRevision", args...)
	if err != nil {
		debug.Log("Command failed: %s", string(stderr))
		return nil, err
	}
	return stdout, nil
}

// Status return the fossil status output.
func (f *Fossil) Status(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := f.captureCmd(ctx, "FossilStatus", "status")
	if err != nil {
		debug.Log("Command failed: %s\n%s", string(stdout), string(stderr))
		return nil, err
	}
	return stdout, nil
}

// Compact will run fossil rebuild.
func (f *Fossil) Compact(ctx context.Context) error {
	return f.Cmd(ctx, "fossilRebuild", "rebuild", "--compress", "--analyze", "--vacuum")
}
