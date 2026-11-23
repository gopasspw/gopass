package root

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	sec := secrets.NewAKV()
	sec.SetPassword("foo")
	_, err = sec.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, rs.Set(ctx, "zab", sec))

	err = rs.Set(ctx, "zab2", sec)
	require.NoError(t, err)
}

func TestSetMountPointRejected(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	require.NoError(t, u.InitStore("test"))
	require.NoError(t, rs.AddMount(ctx, "test", u.StoreDir("test")))

	require.NoError(t, u.InitStore("nested/sub"))
	require.NoError(t, rs.AddMount(ctx, "nested/sub", u.StoreDir("nested/sub")))

	sec := secrets.NewAKV()
	sec.SetPassword("secret")

	// Creating a secret with the exact name of a mount point must fail.
	for _, bad := range []string{"test", "test/", "nested/sub", "nested/sub/"} {
		err := rs.Set(ctx, bad, sec)
		require.Error(t, err, "expected Set(%q) to fail", bad)
		assert.Contains(t, err.Error(), "mount point")
	}

	// Empty / root-only names must also fail.
	for _, bad := range []string{"", "/", ".", "test/.", "test/.."} {
		require.Error(t, rs.Set(ctx, bad, sec), "expected Set(%q) to fail", bad)
	}

	// Ensure no empty-basename secret file (e.g. ".txt" / ".gpg") was created in the mount.
	for _, ext := range []string{".txt", ".gpg", ".age"} {
		_, statErr := os.Stat(filepath.Join(u.StoreDir("test"), ext))
		assert.True(t, os.IsNotExist(statErr), "unexpected %s file created in mount root", ext)
	}

	// Creating secrets inside the mounts or with non-matching names must still work.
	for _, good := range []string{"test/secret", "test2", "nested", "nested/sub/secret"} {
		require.NoError(t, rs.Set(ctx, good, sec), "expected Set(%q) to succeed", good)
	}
}
