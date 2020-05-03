package action

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestShow(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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
	c.Context = ctx

	assert.NoError(t, act.Show(c))
	assert.Equal(t, "secret", buf.String())
	buf.Reset()

	// show --sync foo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf := cli.BoolFlag{
		Name:  "sync",
		Usage: "sync",
	}
	assert.NoError(t, bf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--sync", "foo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Show(c))
	assert.Equal(t, "secret", buf.String())
	buf.Reset()

	// show dir
	assert.NoError(t, act.Store.Set(ctx, "bar/baz", secret.New("123", "---\nbar: zab")))
	buf.Reset()

	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Show(c))
	assert.Equal(t, "bar\n└── baz\n\n", buf.String())
	buf.Reset()

	// show twoliner with safecontent enabled
	ctx = ctxutil.WithShowSafeContent(ctx, true)
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar/baz"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Show(c))
	assert.Equal(t, "---\nbar: zab", buf.String())
	buf.Reset()

	// show foo with safecontent enabled, should error out
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.Error(t, act.Show(c))
	buf.Reset()

	// show foo with safecontent enabled, with the force flag
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, bf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--force", "foo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Show(c))
	assert.Equal(t, "secret", buf.String())
	buf.Reset()

	// show twoliner with safecontent enabled, but with the clip flag, which should copy just the secret
	ctx = ctxutil.WithShowSafeContent(ctx, true)
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	bf = cli.BoolFlag{
		Name:  "clip",
		Usage: "clip",
	}
	assert.NoError(t, bf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--clip", "bar/baz"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Show(c))
	assert.NotContains(t, buf.String(), "123")
	buf.Reset()
}

func TestShowHandleRevision(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.showHandleRevision(ctx, c, "foo", "HEAD"))
	buf.Reset()
}

func TestShowHandleError(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.Error(t, act.showHandleError(ctx, c, "foo", false, fmt.Errorf("test")))
	buf.Reset()
}

func TestShowHandleYAMLError(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	assert.Error(t, act.showHandleYAMLError(ctx, "foo", "bar", fmt.Errorf("test")))
	buf.Reset()
}

func TestShowPrintQR(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	assert.NoError(t, act.showPrintQR(ctx, "foo", "bar"))
	buf.Reset()
}
