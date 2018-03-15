package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestInsert(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// insert bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Insert(ctx, c))

	// insert bar baz
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar", "baz"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Insert(ctx, c))

	// insert baz via stdin
	assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar")))
	buf.Reset()

	assert.NoError(t, act.show(ctx, c, "baz", "", false))
	assert.Equal(t, "foobar\n", buf.String())
	buf.Reset()

	// insert zab#key
	assert.NoError(t, act.insertYAML(ctx, "zab", "key", []byte("foobar")))

}
