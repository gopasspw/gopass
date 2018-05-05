package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestSync(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()

	// default
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Sync(ctx, c))
	buf.Reset()

	// sync --store=root
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.StringFlag{
		Name:  "store",
		Usage: "store",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--store=root"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Sync(ctx, c))
	buf.Reset()
}
