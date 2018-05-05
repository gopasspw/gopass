package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/justwatchcom/gopass/pkg/tree"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestList(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.List(ctx, c))
	want := `gopass
└── foo

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// add foo/bar and list folder foo
	assert.NoError(t, act.Store.Set(ctx, "foo/bar", secret.New("123", "---\nbar: zab")))
	buf.Reset()

	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.List(ctx, c))
	want = `foo
└── bar

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// list --flat foo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.BoolFlag{
		Name:  "flat",
		Usage: "flat",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--flat=true", "foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.List(ctx, c))
	want = `foo/bar
`
	assert.Equal(t, want, buf.String())
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
