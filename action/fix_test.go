package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestFix(t *testing.T) {
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
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// fix
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Fix(ctx, c))

	// fix --force
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force", "true"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Fix(ctx, c))
}
