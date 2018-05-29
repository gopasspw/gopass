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
	"github.com/urfave/cli"
)

func TestCopy(t *testing.T) {
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

	// copy foo bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo", "bar"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Copy(ctx, c))
	buf.Reset()

	// copy foo bar (again, should fail)
	{
		ctx := ctxutil.WithAlwaysYes(ctx, false)
		ctx = ctxutil.WithInteractive(ctx, false)
		assert.Error(t, act.Copy(ctx, c))
		buf.Reset()
	}

	// copy not-found still-not-there
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"not-found", "still-not-there"}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Copy(ctx, c))
	buf.Reset()

	// copy
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Copy(ctx, c))
	buf.Reset()

	// insert bam/baz
	assert.NoError(t, act.insertStdin(ctx, "bam/baz", []byte("foobar"), false))
	assert.NoError(t, act.insertStdin(ctx, "bam/zab", []byte("barfoo"), false))

	// recursive copy: bam -> zab
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bam", "zab"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Copy(ctx, c))
	buf.Reset()

	assert.NoError(t, act.show(ctx, c, "zab/zab", "", false))
	assert.Equal(t, "barfoo\n", buf.String())
	buf.Reset()
}
