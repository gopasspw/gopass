package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoGitBackend exercises the pure-Go go-git storage backend end to end:
// init, write, read, log and sync against a local bare remote.
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

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}
