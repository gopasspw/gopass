package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	require.NoError(t, act.cfg.Set("", "generate.autoclip", "false"))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	color.NoColor = true

	// copy foo bar
	c := gptest.CliCtx(ctx, t, "foo", "bar")
	require.NoError(t, act.Copy(c))
	buf.Reset()

	// copy foo bar (again, should fail)
	{
		ctx := ctxutil.WithAlwaysYes(ctx, false)
		ctx = ctxutil.WithInteractive(ctx, false)
		c.Context = ctx
		require.Error(t, act.Copy(c))
		buf.Reset()
	}

	// copy not-found still-not-there
	c = gptest.CliCtx(ctx, t, "not-found", "still-not-there")
	require.Error(t, act.Copy(c))
	buf.Reset()

	// copy
	c = gptest.CliCtx(ctx, t)
	require.Error(t, act.Copy(c))
	buf.Reset()

	// insert bam/baz
	require.NoError(t, act.insertStdin(ctx, "bam/baz", []byte("foobar"), false))
	require.NoError(t, act.insertStdin(ctx, "bam/zab", []byte("barfoo"), false))

	// recursive copy: bam/ -> zab
	c = gptest.CliCtx(ctx, t, "bam", "zab")
	require.NoError(t, act.Copy(c))
	buf.Reset()

	require.NoError(t, act.List(gptest.CliCtx(ctx, t)))
	want := `gopass
├── bam/
│   ├── baz
│   └── zab
├── bar
├── foo
└── zab/
    ├── baz
    └── zab

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	ctx = ctxutil.WithTerminal(ctx, false)
	require.NoError(t, act.show(ctx, c, "zab/zab", false))
	assert.Equal(t, "barfoo\n", buf.String())
	buf.Reset()
}

func TestCopyGpg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	u := gptest.NewGUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	require.NoError(t, act.cfg.Set("", "generate.autoclip", "false"))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	color.NoColor = true

	// generate foo
	c := gptest.CliCtx(ctx, t, "foo")
	require.NoError(t, act.Generate(c))
	buf.Reset()

	// copy foo bar
	c = gptest.CliCtx(ctx, t, "foo", "bar")
	require.NoError(t, act.Copy(c))
	buf.Reset()

	// copy foo bar (again, should fail)
	{
		ctx := ctxutil.WithAlwaysYes(ctx, false)
		ctx = ctxutil.WithInteractive(ctx, false)
		c.Context = ctx
		require.Error(t, act.Copy(c))
		buf.Reset()
	}

	// copy not-found still-not-there
	c = gptest.CliCtx(ctx, t, "not-found", "still-not-there")
	require.Error(t, act.Copy(c))
	buf.Reset()

	// copy
	c = gptest.CliCtx(ctx, t)
	require.Error(t, act.Copy(c))
	buf.Reset()

	// insert bam/baz
	require.NoError(t, act.insertStdin(ctx, "bam/baz", []byte("foobar"), false))
	require.NoError(t, act.insertStdin(ctx, "bam/zab", []byte("barfoo"), false))

	// recursive copy: bam/ -> zab
	c = gptest.CliCtx(ctx, t, "bam", "zab")
	require.NoError(t, act.Copy(c))
	buf.Reset()

	require.NoError(t, act.List(gptest.CliCtx(ctx, t)))
	want := `gopass
├── bam/
│   ├── baz
│   └── zab
├── bar
├── foo
└── zab/
    ├── baz
    └── zab

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	ctx = ctxutil.WithTerminal(ctx, false)
	require.NoError(t, act.show(WithPasswordOnly(ctx, true), c, "zab/zab", false))
	assert.Equal(t, "barfoo", buf.String())
	buf.Reset()
}
