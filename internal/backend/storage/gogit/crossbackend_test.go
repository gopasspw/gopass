package gogit

import (
	"context"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/storage/gitfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rcsOps is the subset of backend.Storage used by the cross-backend test.
type rcsOps interface {
	Set(ctx context.Context, name string, value []byte) error
	Get(ctx context.Context, name string) ([]byte, error)
	Add(ctx context.Context, files ...string) error
	Commit(ctx context.Context, msg string) error
	Delete(ctx context.Context, name string) error
	Pull(ctx context.Context, remote, branch string) error
}

// TestCrossBackendObjectGraph runs the same sequence of operations against the
// pure-Go backend and the git CLI backend and compares the resulting object
// graph (tree hashes and file modes), not the status text. Commit hashes are
// intentionally not compared since they depend on author/committer time.
//
// The test is skipped when the git binary is unavailable (the CLI side needs
// it).
func TestCrossBackendObjectGraph(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	ctx := testContext()

	ops := func(t *testing.T, g rcsOps) {
		t.Helper()

		require.NoError(t, g.Set(ctx, "a.gpg", []byte("a1")))
		require.NoError(t, g.Set(ctx, "b.gpg", []byte("b1")))
		require.NoError(t, g.Add(ctx, "."))
		require.NoError(t, g.Commit(ctx, "add a and b"))

		require.NoError(t, g.Set(ctx, "a.gpg", []byte("a2")))
		require.NoError(t, g.Add(ctx, "."))
		require.NoError(t, g.Commit(ctx, "modify a"))

		require.NoError(t, g.Delete(ctx, "b.gpg"))
		require.NoError(t, g.Add(ctx, "."))
		require.NoError(t, g.Commit(ctx, "delete b"))
	}

	// pure-Go backend.
	goDir := t.TempDir()
	goGit, err := Init(ctx, goDir, "Test", "test@example.org")
	require.NoError(t, err)
	ops(t, goGit)

	// git CLI backend.
	cliDir := t.TempDir()
	cliGit, err := gitfs.Init(ctx, cliDir, "Test", "test@example.org")
	require.NoError(t, err)
	ops(t, cliGit)

	// compare the final tree hashes.
	goTree := treeHash(t, goDir)
	cliTree := treeHash(t, cliDir)
	assert.Equal(t, cliTree, goTree, "final tree hashes must match between backends")

	// compare the tracked file list and modes.
	assert.Equal(t, lsTree(t, cliDir), lsTree(t, goDir))
}

// TestCrossBackendMergeGraph runs the same non-conflicting merge against both
// backends and compares the resulting tree and the merge commit's parent count.
func TestCrossBackendMergeGraph(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	ctx := testContext()

	// seed a remote with a base commit.
	remoteDir := t.TempDir()
	remote, err := Init(ctx, remoteDir, "Remote", "remote@example.org")
	require.NoError(t, err)
	require.NoError(t, remote.Set(ctx, "base.gpg", []byte("base")))
	require.NoError(t, remote.Add(ctx, "base.gpg"))
	require.NoError(t, remote.Commit(ctx, "base"))

	merge := func(t *testing.T, g rcsOps, dir, tag string) {
		t.Helper()

		// local makes a non-conflicting change.
		require.NoError(t, g.Set(ctx, "local.gpg", []byte("local-"+tag)))
		require.NoError(t, g.Add(ctx, "local.gpg"))
		require.NoError(t, g.Commit(ctx, "local"))

		// remote moves ahead with a non-conflicting change.
		require.NoError(t, remote.Set(ctx, "remote.gpg", []byte("remote-"+tag)))
		require.NoError(t, remote.Add(ctx, "remote.gpg"))
		require.NoError(t, remote.Commit(ctx, "remote"))

		require.NoError(t, g.Pull(ctx, "origin", "master"))

		// both changes must be present.
		got, err := g.Get(ctx, "remote.gpg")
		require.NoError(t, err)
		assert.Equal(t, "remote-"+tag, string(got))
		got, err = g.Get(ctx, "local.gpg")
		require.NoError(t, err)
		assert.Equal(t, "local-"+tag, string(got))

		// the merge commit must have two parents.
		out := gitInDir(t, dir, "rev-list", "--parents", "-n", "1", "HEAD")
		assert.Len(t, strings.Fields(out), 3, "merge commit must have two parents")
	}

	goDir := t.TempDir()
	goGit, err := Clone(ctx, remoteDir, goDir, "Local", "local@example.org")
	require.NoError(t, err)
	merge(t, goGit, goDir, "go")

	cliDir := t.TempDir()
	cliGit, err := gitfs.Clone(ctx, remoteDir, cliDir, "Local", "local@example.org")
	require.NoError(t, err)
	merge(t, cliGit, cliDir, "cli")

	// the set of tracked paths must match between backends.
	assert.Equal(t, trackedPaths(t, cliDir), trackedPaths(t, goDir), "merged file sets must match between backends")
}

// gitInDir runs the git CLI inside dir and returns its stdout. It sets cmd.Dir
// rather than passing `-C dir`: git implements `-C` with a chdir(), and the
// MSYS2 git used on Windows resolves that through a POSIX chdir() that does not
// accept a Windows-style backslash path (e.g. `C:\msys64\tmp\...\001`), so it
// fails with exit status 128 even though the repository is valid. Setting the
// process working directory via cmd.Dir avoids the conversion entirely; the
// gitfs backend uses cmd.Dir for the same reason.
func gitInDir(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	require.NoError(t, err, "git %s in %s failed: %s", strings.Join(args, " "), dir, stderr.String())

	return string(out)
}

// trackedPaths returns the sorted list of paths tracked at HEAD.
func trackedPaths(t *testing.T, dir string) []string {
	t.Helper()

	paths := strings.Fields(gitInDir(t, dir, "ls-tree", "-r", "--name-only", "HEAD"))
	sort.Strings(paths)

	return paths
}

// treeHash returns the tree hash of HEAD in the given repository using the git
// CLI (works for both backends since they share the on-disk format). It uses
// `show --format=%T` rather than `rev-parse HEAD^{tree}` because the MSYS2
// runtime on Windows performs brace expansion on arguments, turning `{tree}`
// into `tree` and making git fail with "ambiguous argument 'HEAD^tree'".
func treeHash(t *testing.T, dir string) string {
	t.Helper()

	return strings.TrimSpace(gitInDir(t, dir, "show", "-s", "--format=%T", "HEAD"))
}

// lsTree returns the `git ls-tree -r HEAD` output for the given repository.
func lsTree(t *testing.T, dir string) string {
	t.Helper()

	return gitInDir(t, dir, "ls-tree", "-r", "HEAD")
}
