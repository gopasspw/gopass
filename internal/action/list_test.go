package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) { //nolint:paralleltest
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	assert.NoError(t, act.List(gptest.CliCtx(ctx, t)))
	want := `gopass
└── foo

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// add foo/bar and list folder foo
	sec := &secrets.Plain{}
	sec.SetPassword("123")
	assert.Error(t, sec.Set("bar", "zab"))
	assert.NoError(t, act.Store.Set(ctx, "foo/bar", sec))
	buf.Reset()

	assert.NoError(t, act.List(gptest.CliCtx(ctx, t, "foo")))
	want = `foo/
└── bar

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list --flat foo
	assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true"}, "foo")))
	want = `foo/bar
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list --folders

	// add more folders and subfolders
	sec = &secrets.Plain{}
	sec.SetPassword("123")
	assert.NoError(t, act.Store.Set(ctx, "foo/zen/bar", sec))
	assert.NoError(t, act.Store.Set(ctx, "foo2/bar2", sec))
	buf.Reset()

	assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true"})))
	want = `foo/
foo/zen/
foo2/
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list not-present
	assert.Error(t, act.List(gptest.CliCtx(ctx, t, "not-present")))
	buf.Reset()
}

func TestListLimit(t *testing.T) { //nolint:paralleltest
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	assert.NoError(t, act.List(gptest.CliCtx(ctx, t)))
	want := `gopass
└── foo

`
	sec := &secrets.Plain{}
	sec.SetPassword("123")
	assert.NoError(t, act.Store.Set(ctx, "foo/bar", sec))
	assert.NoError(t, act.Store.Set(ctx, "foo/zen/baz/bar", sec))
	assert.NoError(t, act.Store.Set(ctx, "foo2/bar2", sec))
	assert.Equal(t, want, buf.String())
	buf.Reset()

	t.Run("folders-limit-0", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true", "limit": "0"})))
		want = `foo/
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit-1", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true", "limit": "1"})))
		want = `foo/
foo/zen/
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit--1", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true", "limit": "-1"})))
		want = `foo/
foo/zen/
foo/zen/baz/
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("flat-limit--1", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true", "limit": "-1"})))
		want = `foo/bar
foo/zen/baz/bar
foo2/bar2
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit-0", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true", "limit": "0"})))
		want = `foo/
foo2/
`
		assert.Equal(t, want, buf.String())
		buf.Reset()
	})

	t.Run("folders-limit-2", func(t *testing.T) { //nolint:paralleltest
		assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"flat": "true", "limit": "2"})))
		want = `foo/bar
foo/zen/baz/
foo2/bar2
`

		assert.Equal(t, want, buf.String())
		buf.Reset()
	})
}

func TestRedirectPager(t *testing.T) { //nolint:paralleltest
	ctx := context.Background()

	var buf *bytes.Buffer
	var subtree *tree.Root

	// no pager
	ctx = ctxutil.WithNoPager(ctx, true)
	so, buf := redirectPager(ctx, subtree)
	assert.Nil(t, buf)
	assert.NotNil(t, so)

	// no term
	ctx = ctxutil.WithNoPager(ctx, false)
	so, buf = redirectPager(ctx, subtree)
	assert.Nil(t, buf)
	assert.NotNil(t, so)
}
