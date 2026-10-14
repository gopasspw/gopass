package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoGitBackend exercises the pure-Go go-git storage backend end to end:
// init, write, read and log.
func TestGoGitBackend(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	// init a store using the pure-Go backend.
	out, err := ts.run("init --crypto=gpgcli --storage=gogit " + keyID)
	require.NoError(t, err, "failed to init gogit store:\n%s", out)
	assert.Contains(t, out, "initialized")

	// the .git directory must exist and be a real git repository.
	require.True(t, isDir(filepath.Join(ts.storeDir("root"), ".git")))

	// write and read a secret.
	out, err = ts.runCmd([]string{ts.Binary, "insert", "--force", "gogit/secret"}, []byte("s3cr3t"))
	require.NoError(t, err, "failed to insert:\n%s", out)

	out, err = ts.run("show -o gogit/secret")
	require.NoError(t, err, "failed to show:\n%s", out)
	assert.Equal(t, "s3cr3t", out)

	// the secret must be committed to the go-git repository.
	out, err = ts.run("git log")
	require.NoError(t, err, "failed to run git log:\n%s", out)
	assert.Contains(t, out, "gogit/secret")
}

// TestGoGitBackendSync verifies that a gogit store can push to and pull from a
// remote via `gopass sync`, including integrating a remote that is ahead.
func TestGoGitBackendSync(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("init --crypto=gpgcli --storage=gogit " + keyID)
	require.NoError(t, err, "failed to init gogit store:\n%s", out)

	// create a bare remote and add it as origin.
	remoteDir := filepath.Join(ts.tempDir, "remote.git")
	require.NoError(t, exec.Command("git", "init", "--bare", remoteDir).Run())
	require.NoError(t, exec.Command("git", "-C", ts.storeDir("root"), "remote", "add", "origin", remoteDir).Run())

	// write a secret and sync.
	out, err = ts.runCmd([]string{ts.Binary, "insert", "--force", "gogit/synced"}, []byte("s3cr3t"))
	require.NoError(t, err, "failed to insert:\n%s", out)

	out, err = ts.run("sync")
	require.NoError(t, err, "failed to sync:\n%s", out)

	// the secret must be present in the bare remote.
	raw, err := exec.Command("git", "-C", remoteDir, "log", "--all", "--name-only").Output()
	require.NoError(t, err)
	assert.Contains(t, string(raw), "gogit/synced")

	// simulate a second clone pushing a change, then sync the first store: the
	// remote is now ahead and the pull must integrate it.
	secondDir := filepath.Join(ts.tempDir, "second")
	require.NoError(t, exec.Command("git", "clone", remoteDir, secondDir).Run())
	require.NoError(t, os.WriteFile(filepath.Join(secondDir, "remote-only.gpg"), []byte("remote"), 0o600))
	require.NoError(t, exec.Command("git", "-C", secondDir, "add", "remote-only.gpg").Run())
	require.NoError(t, exec.Command("git", "-C", secondDir, "-c", "user.name=t", "-c", "user.email=t@e.org", "commit", "-m", "remote-only").Run())
	require.NoError(t, exec.Command("git", "-C", secondDir, "push", "origin", "HEAD:master").Run())

	out, err = ts.run("sync")
	require.NoError(t, err, "failed to sync with remote ahead:\n%s", out)

	_, err = os.Stat(filepath.Join(ts.storeDir("root"), "remote-only.gpg"))
	require.NoError(t, err, "the remote-only file must be integrated by sync")
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}
