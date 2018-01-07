package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/tree"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestList(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, td)
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
