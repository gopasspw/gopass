package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secret"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act, err := newMock(ctx, u)
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
	sec := secret.New()
	sec.Set("password", "123")
	sec.Set("bar", "zab")
	assert.NoError(t, act.Store.Set(ctx, "foo/bar", sec))
	buf.Reset()

	assert.NoError(t, act.List(gptest.CliCtx(ctx, t, "foo")))
	want = `foo
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
	sec = secret.New()
	sec.Set("password", "123")
	assert.NoError(t, act.Store.Set(ctx, "foo/zen/bar", sec))
	assert.NoError(t, act.Store.Set(ctx, "foo2/bar2", sec))
	buf.Reset()

	assert.NoError(t, act.List(gptest.CliCtxWithFlags(ctx, t, map[string]string{"folders": "true"})))
	want = `foo
foo/zen
foo2
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list not-present
	assert.Error(t, act.List(gptest.CliCtx(ctx, t, "not-present")))
	buf.Reset()
}

func TestRedirectPager(t *testing.T) {
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
