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

func TestInsert(t *testing.T) {
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

	// insert zab#key
	assert.NoError(t, act.insertYAML(ctx, "zab", "key", []byte("foobar")))
}
