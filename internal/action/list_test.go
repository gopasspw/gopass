package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	require.NoError(t, act.List(gptest.CliCtx(ctx, t)))
	want := `gopass
└── foo

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// add foo/bar and list folder foo
	sec := secrets.NewAKV()
	sec.SetPassword("123")
	require.NoError(t, sec.Set("bar", "zab"))
	require.NoError(t, act.Store.Set(ctx, "foo/bar", sec))
	buf.Reset()

	require.NoError(t, act.List(gptest.CliCtx(ctx, t, "foo")))
	want = `foo/
└── bar

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list --flat foo
	require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true"}, "foo")))
	want = `foo/bar
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list --folders

	// add more folders and subfolders
	sec = secrets.NewAKV()
	sec.SetPassword("123")
	require.NoError(t, act.Store.Set(ctx, "foo/zen/bar", sec))
	require.NoError(t, act.Store.Set(ctx, "foo2/bar2", sec))
	buf.Reset()

	require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true"})))
	want = `foo/
foo/zen/
foo2/
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// add shadowed entry
	sec = secrets.NewAKV()
	sec.SetPassword("123")
	require.NoError(t, act.Store.Set(ctx, "foo/zen", sec))
	buf.Reset()

	require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true"})))
	want = `foo
foo/bar
foo/zen
foo/zen/bar
foo2/bar2
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	require.NoError(t, act.List(gptest.CliCtx(ctx, t, "foo")))
	want = `foo/
├── bar
└── zen/ (shadowed)
    └── bar

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list not-present
	require.Error(t, act.List(gptest.CliCtx(ctx, t, "not-present")))
	buf.Reset()
}

func TestListLimit(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	require.NoError(t, act.List(gptest.CliCtx(ctx, t)))
	want := `gopass
└── foo

`
	sec := secrets.NewAKV()
	sec.SetPassword("123")
	require.NoError(t, act.Store.Set(ctx, "foo/bar", sec))
	require.NoError(t, act.Store.Set(ctx, "foo/zen/baz/bar", sec))
	require.NoError(t, act.Store.Set(ctx, "foo2/bar2", sec))
	assert.Equal(t, want, buf.String())
	buf.Reset()

	t.Run("folders-limit-0", func(t *testing.T) {
		require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true", "limit": "0"})))
		want = `foo/
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit-1", func(t *testing.T) {
		require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true", "limit": "1"})))
		want = `foo/
foo/zen/
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit--1", func(t *testing.T) {
		require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true", "limit": "-1"})))
		want = `foo/
foo/zen/
foo/zen/baz/
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("flat-limit--1", func(t *testing.T) {
		require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true", "limit": "-1"})))
		want = `foo
foo/bar
foo/zen/baz/bar
foo2/bar2
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit-0", func(t *testing.T) {
		require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true", "limit": "0"})))
		want = `foo
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit-2", func(t *testing.T) {
		require.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true", "limit": "2"})))
		want = `foo
foo/bar
foo/zen/baz/
foo2/bar2
`

		assert.Equal(t, want, buf.String())
		buf.Reset()
	})
}

func TestRedirectPager(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())

	var buf *bytes.Buffer
	var subtree *tree.Root

	cfg := config.NewNoWrites()
	ctx = cfg.WithConfig(ctx)

	// no pager
	require.NoError(t, cfg.Set("", "core.nopager", "true"))
	so, buf := redirectPager(ctx, subtree)
	assert.Nil(t, buf)
	assert.NotNil(t, so)

	// no term
	require.NoError(t, cfg.Set("", "core.nopager", "false"))
	so, buf = redirectPager(ctx, subtree)
	assert.Nil(t, buf)
	assert.NotNil(t, so)
}
