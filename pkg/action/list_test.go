package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/pkg/tree"
	"github.com/gopasspw/gopass/tests/gptest"

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

	assert.NoError(t, act.List(clictx(ctx, t)))
	want := `gopass
└── foo

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// add foo/bar and list folder foo
	assert.NoError(t, act.Store.Set(ctx, "foo/bar", secret.New("123", "---\nbar: zab")))
	buf.Reset()

	assert.NoError(t, act.List(clictx(ctx, t, "foo")))
	want = `foo
└── bar

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list --flat foo
	assert.NoError(t, act.List(clictxf(ctx, t, map[string]string{"flat": "true"}, "foo")))
	want = `foo/bar
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list --folders

	// add more folders and subfolders
	assert.NoError(t, act.Store.Set(ctx, "foo/zen/bar", secret.New("123", "")))
	assert.NoError(t, act.Store.Set(ctx, "foo2/bar2", secret.New("123", "")))
	buf.Reset()

	assert.NoError(t, act.List(clictxf(ctx, t, map[string]string{"folders": "true"})))
	want = `foo
foo/zen
foo2
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list not-present
	assert.Error(t, act.List(clictx(ctx, t, "not-present")))
	buf.Reset()
}

func TestRedirectPager(t *testing.T) {
	ctx := context.Background()

	var buf *bytes.Buffer
	var subtree tree.Tree

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
