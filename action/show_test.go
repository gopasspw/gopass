package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestShow(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// show foo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Show(ctx, c))
	assert.Equal(t, "0xDEADBEEF", buf.String())
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
	assert.Equal(t, "0xDEADBEEF", buf.String())
	buf.Reset()
}
