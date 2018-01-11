package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestShow(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// show foo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Show(ctx, c))
	assert.Equal(t, "secret", buf.String())
	buf.Reset()

	// show --sync foo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.BoolFlag{
		Name:  "sync",
		Usage: "sync",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--sync", "foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Show(ctx, c))
	assert.Equal(t, "secret", buf.String())
	buf.Reset()

	// show dir
	assert.NoError(t, act.Store.Set(ctx, "bar/baz", secret.New("123", "---\nbar: zab")))
	buf.Reset()

	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Show(ctx, c))
	assert.Equal(t, "bar\n└── baz\n\n", buf.String())
	buf.Reset()
}
