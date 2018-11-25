package root

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMove(t *testing.T) {
	u := gptest.NewUnitTester(t)
	u.Entries = []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}
	require.NoError(t, u.InitStore(""))
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	assert.NoError(t, rs.Delete(ctx, "foo"))

	// Initial state:
	// foo/bar
	// foo/baz
	// misc/zab
	entries, err := rs.List(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}, entries)
	// -> move foo misc => ERROR: foo is a directory
	assert.Error(t, rs.Move(ctx, "foo", "misc"))
	// -> move foo/ misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Move(ctx, "foo/", "misc/zab"))
	// -> move foo/ misc => OK
	assert.NoError(t, rs.Move(ctx, "foo/", "misc"))
	// New state:
	// misc/foo/bar
	// misc/foo/baz
	// misc/zab
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, []string{
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)
	// -> move misc/foo/ bar/ => OK
	assert.NoError(t, rs.Move(ctx, "misc/foo/", "bar/"))
	// New state:
	// bar/foo/bar
	// bar/foo/baz
	// misc/zab
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/foo/bar",
		"bar/foo/baz",
		"misc/zab",
	}, entries)
	// -> move misc/zab bar/foo/zab => OK
	assert.NoError(t, rs.Move(ctx, "misc/zab", "bar/foo/zab"))
	// New state:
	// bar/foo/bar
	// bar/foo/baz
	// bar/foo/zab
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/foo/bar",
		"bar/foo/baz",
		"bar/foo/zab",
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
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	assert.NoError(t, rs.Delete(ctx, "foo"))

	// Initial state:
	// foo/bar
	// foo/baz
	// misc/zab
	entries, err := rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}, entries)
	// -> move foo misc => ERROR: foo is a directory
	assert.Error(t, rs.Copy(ctx, "foo", "misc"))
	// -> move foo/ misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Copy(ctx, "foo/", "misc/zab"))
	// -> move foo/ misc => OK
	assert.NoError(t, rs.Copy(ctx, "foo/", "misc"))
	// New state:
	// misc/foo/bar
	// misc/foo/baz
	// misc/zab
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"foo/bar",
		"foo/baz",
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)
	// -> move misc/foo/ bar/ => OK
	assert.NoError(t, rs.Copy(ctx, "misc/foo/", "bar/"))
	// New state:
	// bar/foo/bar
	// bar/foo/baz
	// misc/zab
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/foo/bar",
		"bar/foo/baz",
		"foo/bar",
		"foo/baz",
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)
	// -> move misc/zab bar/foo/zab => OK
	assert.NoError(t, rs.Copy(ctx, "misc/zab", "bar/foo/zab"))
	// New state:
	// bar/foo/bar
	// bar/foo/baz
	// bar/foo/zab
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/foo/bar",
		"bar/foo/baz",
		"bar/foo/zab",
		"foo/bar",
		"foo/baz",
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)
}
