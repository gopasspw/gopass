package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	for _, mount := range []string{"m1", "m2"} {
		out, err := ts.run("init --store " + mount + " --path " + ts.storeDir(mount) + " --storage=fs " + keyID)
		require.NoError(t, err, "failed to init mount %s: %s", mount, out)
		ts.initSecrets(mount + "/")
	}

	rootBazBefore, err := os.ReadFile(filepath.Join(ts.storeDir("root"), "baz.gpg"))
	require.NoError(t, err)
	m1BazBefore, err := os.ReadFile(filepath.Join(ts.storeDir("m1"), "baz.gpg"))
	require.NoError(t, err)
	m2BazBefore, err := os.ReadFile(filepath.Join(ts.storeDir("m2"), "baz.gpg"))
	require.NoError(t, err)

	out, err := ts.run("fsck --store=m1 --decrypt")
	require.NoError(t, err, "failed to fsck --store=m1: %s", out)
	assert.Contains(t, out, "[m1] Re-encrypting")
	assert.NotContains(t, out, "[m2]")
	assert.NotContains(t, out, "[] Re-encrypting")

	rootBazAfter, err := os.ReadFile(filepath.Join(ts.storeDir("root"), "baz.gpg"))
	require.NoError(t, err)
	m1BazAfter, err := os.ReadFile(filepath.Join(ts.storeDir("m1"), "baz.gpg"))
	require.NoError(t, err)
	m2BazAfter, err := os.ReadFile(filepath.Join(ts.storeDir("m2"), "baz.gpg"))
	require.NoError(t, err)

	assert.Equal(t, rootBazBefore, rootBazAfter, "root store secret must not be re-encrypted")
	assert.Equal(t, m2BazBefore, m2BazAfter, "unrelated mount m2 secret must not be re-encrypted")
	assert.NotEqual(t, m1BazBefore, m1BazAfter, "target mount m1 secret should be re-encrypted")

	out, err = ts.run("fsck --store=nonexistent")
	require.Error(t, err)
	assert.Contains(t, out, "nonexistent")
}
