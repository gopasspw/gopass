package root

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveShadow(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"old/www/foo",
		"old/www/bar",
	}

	require.NoError(t, u.InitStore(""))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Delete(ctx, "foo"))

	// -> move old/www/foo www/dir/foo => OK
	require.NoError(t, rs.Move(ctx, "old/www/foo", "www/dir/foo"))
	entries, err := rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"old/www/bar",
		"www/dir/foo",
	}, entries)

	// -> move old/www/bar www/ => OK
	require.NoError(t, rs.Move(ctx, "old/www/bar", "www/"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"www/bar",
		"www/dir/foo",
	}, entries)
}

func TestMove(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}
	require.NoError(t, u.InitStore(""))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Delete(ctx, "foo"))

	// Initial state:
	entries, err := rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}, entries)

	// -> move foo/ misc/zab => ERROR: misc/zab is a file
	require.Error(t, rs.Move(ctx, "foo/", "misc/zab"))

	// -> move foo misc/zab => ERROR: misc/zab is a file
	require.Error(t, rs.Move(ctx, "foo", "misc/zab"))

	// -> move foo misc => OK
	require.NoError(t, rs.Move(ctx, "foo", "misc"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)

	// -> move misc/foo bar/ => OK
	require.NoError(t, rs.Move(ctx, "misc/foo", "bar/"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/bar",
		"bar/baz",
		"misc/zab",
	}, entries)

	// -> move misc/zab bar/foo/zab => OK
	require.NoError(t, rs.Move(ctx, "misc/zab", "bar/foo/zab"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/bar",
		"bar/baz",
		"bar/foo/zab",
	}, entries)

	// -> move bar/foo/ baz => OK
	require.NoError(t, rs.Move(ctx, "bar/foo/", "baz"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/bar",
		"bar/baz",
		"baz/zab",
	}, entries)

	// -> move baz/ boz/ => OK
	require.NoError(t, rs.Move(ctx, "baz/", "boz/"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/bar",
		"bar/baz",
		"boz/zab",
	}, entries)

	// this fails if empty directories are not removed, because 'bar' and 'baz'
	// were directories in the root folder.
	// -> move boz/ / => OK
	require.NoError(t, rs.Move(ctx, "boz/", "."))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/bar",
		"bar/baz",
		"zab",
	}, entries)
}

func TestUnixMvSemantics(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"a/f1",
		"a/f2",
		"b/f3",
	}
	require.NoError(t, u.InitStore(""))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Delete(ctx, "foo"))

	// Initial state:
	entries, err := rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"a/f1",
		"a/f2",
		"b/f3",
	}, entries)

	// -> move a b => Move a below b
	require.NoError(t, rs.Move(ctx, "a", "b"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"b/a/f1",
		"b/a/f2",
		"b/f3",
	}, entries)
}

func TestRegression2079(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"comm/test",
		"comm/test2",
		"communication/t1",
	}
	require.NoError(t, u.InitStore(""))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Delete(ctx, "foo"))

	// Initial state:
	entries, err := rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"comm/test",
		"comm/test2",
		"communication/t1",
	}, entries)

	// -> move comm email => Rename comm to email
	require.NoError(t, rs.Move(ctx, "comm", "email"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"communication/t1",
		"email/test",
		"email/test2",
	}, entries)
}

func TestCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}
	require.NoError(t, u.InitStore(""))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Delete(ctx, "foo"))

	// Initial state:
	t.Run("initial state", func(t *testing.T) {
		entries, err := rs.List(ctx, tree.INF)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"foo/bar",
			"foo/baz",
			"misc/zab",
		}, entries)
	})

	// -> copy foo/ misc/zab => ERROR: misc/zab is a file
	require.Error(t, rs.Copy(ctx, "foo/", "misc/zab"))
	// -> copy foo misc/zab => ERROR: misc/zab is a file
	require.Error(t, rs.Copy(ctx, "foo", "misc/zab"))

	// -> copy foo/ misc => OK
	t.Run("copy foo misc", func(t *testing.T) {
		require.NoError(t, rs.Copy(ctx, "foo", "misc"))
		entries, err := rs.List(ctx, tree.INF)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"foo/bar",
			"foo/baz",
			"misc/foo/bar",
			"misc/foo/baz",
			"misc/zab",
		}, entries)
	})

	// -> copy misc/foo/ bar/ => OK
	t.Run("copy misc/foo/ bar/", func(t *testing.T) {
		require.NoError(t, rs.Copy(ctx, "misc/foo/", "bar/"))
		entries, err := rs.List(ctx, tree.INF)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"bar/bar",
			"bar/baz",
			"foo/bar",
			"foo/baz",
			"misc/foo/bar",
			"misc/foo/baz",
			"misc/zab",
		}, entries)
	})

	// -> copy misc/zab bar/foo/zab => OK
	t.Run("copy misc/zab bar/foo/zab", func(t *testing.T) {
		require.NoError(t, rs.Copy(ctx, "misc/zab", "bar/foo/zab"))
		entries, err := rs.List(ctx, tree.INF)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"bar/bar",
			"bar/baz",
			"bar/foo/zab",
			"foo/bar",
			"foo/baz",
			"misc/foo/bar",
			"misc/foo/baz",
			"misc/zab",
		}, entries)
	})
}

func TestMoveSelf(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"foo/bar/example",
	}
	require.NoError(t, u.InitStore(""))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	// Initial state:
	t.Run("initial state", func(t *testing.T) {
		entries, err := rs.List(ctx, tree.INF)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"foo",
			"foo/bar/example",
		}, entries)
	})

	// -> move foo/bar/example foo/bar -> no op
	t.Run("move self", func(t *testing.T) {
		require.Error(t, rs.Move(ctx, "foo/bar/example", "foo/bar"))
		entries, err := rs.List(ctx, tree.INF)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"foo",
			"foo/bar/example",
		}, entries)
	})
}

func TestComputeMoveDestination(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		src      string
		from     string
		to       string
		srcIsDir bool
		dstIsDir bool
		dst      string
	}{
		{
			name: "rename file a to file b", // mv a b
			src:  "a",
			from: "a",
			to:   "b",
			dst:  "b",
		},
		{
			name:     "rename dir a to dir b (#2079)", // mv comm email
			src:      "comm/test",
			from:     "comm",
			to:       "email",
			dst:      "email/test",
			srcIsDir: true,
			dstIsDir: false,
		},
		{
			name:     "rename dir a to dir b (existing dir)", // mv a b
			src:      "a/f1",
			from:     "a",
			to:       "b",
			dst:      "b/a/f1",
			srcIsDir: true,
			dstIsDir: true,
		},
		{
			name:     "move up", // mv a/b/c c
			src:      "a/b/c/f1",
			from:     "a/b/c",
			to:       "c",
			dst:      "c/f1",
			srcIsDir: true,
		},
		{
			name:     "move fully up", // mv a/ .
			src:      "a/f1",
			from:     "a/",
			to:       ".",
			dst:      "f1",
			srcIsDir: true,
			dstIsDir: true,
		},
		{
			name:     "old www", // mv old/www/bar www/
			src:      "old/www/bar",
			from:     "old/www/bar",
			to:       "www",
			dst:      "www/bar",
			srcIsDir: false,
			dstIsDir: true,
		},
		{
			name:     "one level up", // mv foo/bar/example foo/bar
			src:      "foo/bar/example",
			from:     "foo/bar",
			to:       "foo/bar",
			dst:      "foo/bar/example",
			srcIsDir: false,
			dstIsDir: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dst := computeMoveDestination(tc.src, tc.from, tc.to, tc.srcIsDir, tc.dstIsDir)
			assert.Equal(t, tc.dst, dst, tc.name)
		})
	}
}

func TestRegression892(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"some/example",
		"some/example/test2",
		"communication/t1",
	}
	require.NoError(t, u.InitStore(""))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Delete(ctx, "foo"))

	// Initial state:
	entries, err := rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"communication/t1",
		"some/example",
		"some/example/test2",
	}, entries)

	// -> move comm email => Rename comm to email
	require.NoError(t, rs.Move(ctx, "some/example", "some/example/test1"))
	entries, err = rs.List(ctx, tree.INF)
	require.NoError(t, err)
	require.Equal(t, []string{
		"communication/t1",
		"some/example/test1",
		"some/example/test2",
	}, entries)
}

func TestMoveInMountedStore(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	// create a mount
	require.NoError(t, u.InitStore("m7an"))
	mountDir := u.StoreDir("m7an")
	require.NoError(t, rs.AddMount(ctx, "m7an", mountDir))
	sec := secrets.New()
	sec.SetPassword("foo")
	require.NoError(t, rs.Set(ctx, "m7an/www/hostprvdr.de/hostprvdr@m7an.de", sec))

	// move the secret
	require.NoError(t, rs.Move(ctx, "m7an/www/hostprvdr.de/hostprvdr@m7an.de", "m7an/www/hostprvdr.de/meinhostprvdr@m7an.de"))

	// check if the secret was moved correctly
	_, err = rs.Get(ctx, "m7an/www/hostprvdr.de/meinhostprvdr@m7an.de")
	require.NoError(t, err)

	// check that the old secret is gone
	_, err = rs.Get(ctx, "m7an/www/hostprvdr.de/hostprvdr@m7an.de")
	require.Error(t, err)

	// check that no extra directory was created
	_, err = rs.Get(ctx, "m7an/m7an/www/hostprvdr.de/meinhostprvdr@m7an.de")
	require.Error(t, err)
}

// TestCrossStoreMoveReencrypts is a regression test for
// https://github.com/gopasspw/gopass/issues/3359.
// Moving a secret from one store to another must decrypt it and re-encrypt
// it for the destination store's recipients, not merely copy the raw
// ciphertext file.
func TestCrossStoreMoveReencrypts(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	// Write a secret in the root store.
	sec := secrets.New()
	sec.SetPassword("secret-password")
	require.NoError(t, rs.Set(ctx, "root-secret", sec))

	// Create and mount a second store.
	require.NoError(t, u.InitStore("sub"))
	require.NoError(t, rs.AddMount(ctx, "sub", u.StoreDir("sub")))

	// Move the secret from the root store to the mounted sub-store.
	require.NoError(t, rs.Move(ctx, "root-secret", "sub/root-secret"))

	// The secret must be readable from the destination store.
	got, err := rs.Get(ctx, "sub/root-secret")
	require.NoError(t, err)
	require.Equal(t, "secret-password", got.Password())

	// The secret must no longer exist in the source store.
	_, err = rs.Get(ctx, "root-secret")
	require.Error(t, err)
}

// TestCrossStoreCopyReencrypts is a regression test for
// https://github.com/gopasspw/gopass/issues/3359.
// Copying a secret from one store to another must also go through Get+Set
// to re-encrypt for the destination store's recipients.
func TestCrossStoreCopyReencrypts(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	// Write a secret in the root store.
	sec := secrets.New()
	sec.SetPassword("secret-password")
	require.NoError(t, rs.Set(ctx, "root-secret", sec))

	// Create and mount a second store.
	require.NoError(t, u.InitStore("sub"))
	require.NoError(t, rs.AddMount(ctx, "sub", u.StoreDir("sub")))

	// Copy the secret from the root store to the mounted sub-store.
	require.NoError(t, rs.Copy(ctx, "root-secret", "sub/root-secret"))

	// The secret must be readable from the destination store.
	got, err := rs.Get(ctx, "sub/root-secret")
	require.NoError(t, err)
	require.Equal(t, "secret-password", got.Password())

	// The original must still exist in the source store.
	orig, err := rs.Get(ctx, "root-secret")
	require.NoError(t, err)
	require.Equal(t, "secret-password", orig.Password())
}

// TestCrossStoreMoveDirectory is a regression test for
// https://github.com/gopasspw/gopass/issues/3359.
// All secrets in a directory moved across stores must be re-encrypted.
func TestCrossStoreMoveDirectory(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"folder/a",
		"folder/b",
		"folder/c",
	}
	require.NoError(t, u.InitStore(""))

	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NoError(t, rs.Delete(ctx, "foo"))

	// Create and mount a second store.
	require.NoError(t, u.InitStore("sub"))
	require.NoError(t, rs.AddMount(ctx, "sub", u.StoreDir("sub")))

	// Move the whole folder to the mounted sub-store.
	require.NoError(t, rs.Move(ctx, "folder", "sub/folder"))

	// Every secret must be readable at the destination.
	for _, name := range []string{"sub/folder/a", "sub/folder/b", "sub/folder/c"} {
		_, err := rs.Get(ctx, name)
		require.NoError(t, err, "expected %q to be readable after cross-store move", name)
	}

	// No secret must remain at the source paths.
	_, err = rs.Get(ctx, "folder/a")
	require.Error(t, err)
	_, err = rs.Get(ctx, "folder/b")
	require.Error(t, err)
	_, err = rs.Get(ctx, "folder/c")
	require.Error(t, err)
}
