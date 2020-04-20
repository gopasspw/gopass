package root

import (
	"context"
	"path/filepath"
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
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("misc", "zab"),
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
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("misc", "zab"),
	}, entries)
	// -> move foo/ misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Move(ctx, "foo"+sep, filepath.Join("misc", "zab")))
	// -> move foo misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Move(ctx, "foo", filepath.Join("misc", "zab")))

	// -> move foo misc => OK
	assert.NoError(t, rs.Move(ctx, "foo", "misc"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, []string{
		filepath.Join("misc", "foo", "bar"),
		filepath.Join("misc", "foo", "baz"),
		filepath.Join("misc", "zab"),
	}, entries)

	// -> move misc/foo bar/ => OK
	assert.NoError(t, rs.Move(ctx, filepath.Join("misc", "foo"), "bar"+sep))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("bar", "foo", "bar"),
		filepath.Join("bar", "foo", "baz"),
		filepath.Join("misc", "zab"),
	}, entries)

	// -> move misc/zab bar/foo/zab => OK
	assert.NoError(t, rs.Move(ctx, "misc/zab", "bar/foo/zab"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("bar", "foo", "bar"),
		filepath.Join("bar", "foo", "baz"),
		filepath.Join("bar", "foo", "zab"),
	}, entries)

	// -> move bar/foo/ baz => OK
	assert.NoError(t, rs.Move(ctx, filepath.Join("bar", "foo")+sep, "baz"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("baz", "bar"),
		filepath.Join("baz", "baz"),
		filepath.Join("baz", "zab"),
	}, entries)

	// -> move baz/ boz/ => OK
	assert.NoError(t, rs.Move(ctx, "baz"+sep, "boz"+sep))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("boz", "bar"),
		filepath.Join("boz", "baz"),
		filepath.Join("boz", "zab"),
	}, entries)

	// this fails if empty directories are not removed, because 'bar' and 'baz' were directories in the root folder
	// -> move boz/ / => OK
	assert.NoError(t, rs.Move(ctx, "boz"+sep, sep))
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
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("misc", "zab"),
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
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("misc", "zab"),
	}, entries)
	// -> copy foo/ misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Copy(ctx, "foo"+sep, filepath.Join("misc", "zab")))
	// -> copy foo misc/zab => ERROR: misc/zab is a file
	assert.Error(t, rs.Copy(ctx, "foo", filepath.Join("misc", "zab")))

	// -> copy foo/ misc => OK
	assert.NoError(t, rs.Copy(ctx, "foo", "misc"))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("misc", "foo", "bar"),
		filepath.Join("misc", "foo", "baz"),
		filepath.Join("misc", "zab"),
	}, entries)

	// -> copy misc/foo/ bar/ => OK
	assert.NoError(t, rs.Copy(ctx, filepath.Join("misc", "foo")+sep, "bar"+sep))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("bar", "bar"),
		filepath.Join("bar", "baz"),
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("misc", "foo", "bar"),
		filepath.Join("misc", "foo", "baz"),
		filepath.Join("misc", "zab"),
	}, entries)

	// -> copy misc/zab bar/foo/zab => OK
	assert.NoError(t, rs.Copy(ctx, filepath.Join("misc", "zab"), filepath.Join("bar", "foo", "zab")))
	entries, err = rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("bar", "foo", "zab"),
		filepath.Join("bar", "bar"),
		filepath.Join("bar", "baz"),
		filepath.Join("foo", "bar"),
		filepath.Join("foo", "baz"),
		filepath.Join("misc", "foo", "bar"),
		filepath.Join("misc", "foo", "baz"),
		filepath.Join("misc", "zab"),
	}, entries)
}
