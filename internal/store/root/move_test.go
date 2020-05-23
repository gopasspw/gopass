package root

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

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
	entries, err := rs.List(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}, entries)
	// -> move foo/ misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Move(ctx, "foo/", "misc/zab"))
	// -> move foo misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Move(ctx, "foo", "misc/zab"))

	// -> move foo misc => OK
	assert.NoError(t, rs.Move(ctx, "foo", "misc"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, []string{
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)

	// -> move misc/foo bar/ => OK
	assert.NoError(t, rs.Move(ctx, "misc/foo", "bar/"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/foo/bar",
		"bar/foo/baz",
		"misc/zab",
	}, entries)

	// -> move misc/zab bar/foo/zab => OK
	assert.NoError(t, rs.Move(ctx, "misc/zab", "bar/foo/zab"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/foo/bar",
		"bar/foo/baz",
		"bar/foo/zab",
	}, entries)

	// -> move bar/foo/ baz => OK
	assert.NoError(t, rs.Move(ctx, "bar/foo/", "baz"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"baz/bar",
		"baz/baz",
		"baz/zab",
	}, entries)

	// -> move baz/ boz/ => OK
	assert.NoError(t, rs.Move(ctx, "baz/", "boz/"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"boz/bar",
		"boz/baz",
		"boz/zab",
	}, entries)

	// this fails if empty directories are not removed, because 'bar' and 'baz' were directories in the root folder
	// -> move boz/ / => OK
	assert.NoError(t, rs.Move(ctx, "boz/", "/"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar",
		"baz",
		"zab",
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
	entries, err := rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"foo/bar",
		"foo/baz",
		"misc/zab",
	}, entries)
	// -> copy foo/ misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Copy(ctx, "foo/", "misc/zab"))
	// -> copy foo misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Copy(ctx, "foo", "misc/zab"))

	// -> copy foo/ misc => OK
	assert.NoError(t, rs.Copy(ctx, "foo", "misc"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"foo/bar",
		"foo/baz",
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)

	// -> copy misc/foo/ bar/ => OK
	assert.NoError(t, rs.Copy(ctx, "misc/foo/", "bar/"))
	entries, err = rs.List(ctx, 0)
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

	// -> copy misc/zab bar/foo/zab => OK
	assert.NoError(t, rs.Copy(ctx, "misc/zab", "bar/foo/zab"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"bar/foo/zab",
		"bar/bar",
		"bar/baz",
		"foo/bar",
		"foo/baz",
		"misc/foo/bar",
		"misc/foo/baz",
		"misc/zab",
	}, entries)
}
