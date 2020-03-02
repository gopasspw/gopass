package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestFind(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	app := cli.NewApp()

	// find
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	if err := act.Find(ctx, c); err == nil || err.Error() != "Usage: action.test find <NEEDLE>" {
		t.Errorf("Should fail: %s", err)
	}

	// find fo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"fo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Find(ctx, c))
	assert.Equal(t, "Found exact match in 'foo'\nsecret", strings.TrimSpace(buf.String()))
	buf.Reset()

	// testing the safecontent case
	ctx = ctxutil.WithShowSafeContent(ctx, true)
	assert.NoError(t, act.Find(ctx, c))
	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in 'foo'")
	assert.Contains(t, out, "with -f")
	assert.Contains(t, out, "Copying password instead.")
	buf.Reset()

	// testing with the clip flag set
	bf := cli.BoolFlag{
		Name:  "clip",
		Usage: "clip",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"-clip", "fo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Find(ctx, c))
	out = strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in 'foo'")
	assert.NotContains(t, out, "Copying password instead.")
	buf.Reset()

	// safecontent case with force flag set
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"-force", "fo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Find(ctx, c))
	out = strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in 'foo'\nsecret")
	buf.Reset()

	// stopping with the safecontent tests
	ctx = ctxutil.WithShowSafeContent(ctx, false)

	// find yo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"yo"}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Find(ctx, c))
	buf.Reset()

	// add some secrets
	assert.NoError(t, act.Store.Set(ctx, "bar/baz", secret.New("foo", "bar")))
	assert.NoError(t, act.Store.Set(ctx, "bar/zab", secret.New("foo", "bar")))
	buf.Reset()

	// find bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Find(ctx, c))
	assert.Equal(t, "bar/baz\nbar/zab", strings.TrimSpace(buf.String()))
	buf.Reset()
}
