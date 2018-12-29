package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestInsert(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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
	assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar"), false))
	buf.Reset()

	assert.NoError(t, act.show(ctx, c, "baz", "", false))
	assert.Equal(t, "foobar", buf.String())
	buf.Reset()

	// insert zab#key
	assert.NoError(t, act.insertYAML(ctx, "zab", "key", []byte("foobar"), nil))

	// insert --multiline foo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.BoolFlag{
		Name:  "multiline",
		Usage: "multiline",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--multiline=true", "bar", "baz"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Insert(ctx, c))
}
