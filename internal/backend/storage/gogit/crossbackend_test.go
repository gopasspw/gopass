package gogit

import (
	"context"
	"os/exec"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/storage/gitfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rcsOps is the subset of backend.Storage used by the cross-backend test.
type rcsOps interface {
	Set(ctx context.Context, name string, value []byte) error
	Add(ctx context.Context, files ...string) error
	Commit(ctx context.Context, msg string) error
	Delete(ctx context.Context, name string) error
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

// treeHash returns the tree hash of HEAD in the given repository using the git
// CLI (works for both backends since they share the on-disk format).
func treeHash(t *testing.T, dir string) string {
	t.Helper()

	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD^{tree}").Output()
	require.NoError(t, err)

	return string(out)
}

// lsTree returns the `git ls-tree -r HEAD` output for the given repository.
func lsTree(t *testing.T, dir string) string {
	t.Helper()

	out, err := exec.Command("git", "-C", dir, "ls-tree", "-r", "HEAD").Output()
	require.NoError(t, err)

	return string(out)
}
