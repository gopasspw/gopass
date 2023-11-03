package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowMulti(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	// first add another entry in a subdir
	sec := secrets.NewAKV()
	sec.SetPassword("123")
	require.NoError(t, sec.Set("bar", "zab"))
	require.NoError(t, act.Store.Set(ctx, "bar/baz", sec))
	buf.Reset()

	t.Run("show foo", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtx(ctx, t, "foo")
		require.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "secret")
	})

	t.Run("show --sync foo", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"sync": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "secret")
	})

	t.Run("show dir", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "bar")
		require.NoError(t, act.Show(c))
		assert.Equal(t, "bar/\n└── baz\n\n", buf.String())
		buf.Reset()
	})

	require.NoError(t, act.cfg.Set("", "show.safecontent", "true"))

	t.Run("show twoliner with safecontent enabled", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "bar/baz")

		require.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "bar: zab")
		assert.NotContains(t, buf.String(), "password: ***")
		assert.NotContains(t, buf.String(), "123")
		buf.Reset()
	})

	t.Run("show foo with safecontent enabled, should error out", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "foo")
		require.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "secret")
		buf.Reset()
	})

	t.Run("show foo with safecontent enabled, with the force flag", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"unsafe": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "secret")
		buf.Reset()
	})

	t.Run("show twoliner with safecontent enabled, but with the clip flag, which should copy just the secret", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "bar/baz")

		require.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "123")
		buf.Reset()
	})

	t.Run("show entry with unsafe keys", func(t *testing.T) {
		sec := secrets.NewAKV()
		sec.SetPassword("123")
		require.NoError(t, sec.Set("bar", "zab"))
		require.NoError(t, sec.Set("foo", "baz"))
		require.NoError(t, sec.Set("hello", "world"))
		require.NoError(t, sec.Set("unsafe-keys", "foo, bar"))
		require.NoError(t, act.Store.Set(ctx, "unsafe/keys", sec))
		buf.Reset()

		c := gptest.CliCtx(ctx, t, "unsafe/keys")
		require.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "*****")
		assert.NotContains(t, buf.String(), "zab")
		assert.NotContains(t, buf.String(), "baz")
		buf.Reset()
	})

	t.Run("show twoliner with safecontent enabled", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "bar/baz")

		require.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "bar: zab")
		assert.NotContains(t, buf.String(), "password: ***")
		assert.NotContains(t, buf.String(), "123")
		buf.Reset()
	})

	t.Run("show twoliner with safecontent enabled", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "bar/baz")

		require.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "bar: zab")
		// password should not show up neither be obstructed
		assert.NotContains(t, buf.String(), "123")
		assert.NotContains(t, buf.String(), "***")
		buf.Reset()
	})

	require.NoError(t, act.cfg.Set("", "show.safecontent", "false"))

	t.Run("show key ", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "bar/baz", "bar")

		require.NoError(t, act.Show(c))
		assert.Equal(t, "zab", buf.String())
		buf.Reset()
	})

	t.Run("show nonexisting key", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "bar/baz", "nonexisting")

		require.Error(t, act.Show(c))
		buf.Reset()
	})

	t.Run("show keys with mixed case", func(t *testing.T) {
		require.NoError(t, act.insertStdin(ctx, "baz2", []byte("foobar\nOther: meh\nuser: name\nbody text"), false))
		buf.Reset()

		c := gptest.CliCtx(ctx, t, "baz2", "Other")
		require.NoError(t, act.Show(c))
		assert.Equal(t, "meh", buf.String())
		buf.Reset()
	})

	t.Run("show value with format strings", func(t *testing.T) {
		pw := "some-chars-are-odd-%s-%p-%q"

		require.NoError(t, act.insertStdin(ctx, "printf", []byte(pw), false))
		buf.Reset()

		c := gptest.CliCtx(ctx, t, "printf")
		require.NoError(t, act.Show(c))
		assert.Equal(t, pw+"\n", buf.String())
		assert.NotContains(t, buf.String(), "MISSING")
		buf.Reset()
	})
}

func TestShowAutoClip(t *testing.T) {
	// make sure we consistently get the unsupported error message
	ov := clipboard.Unsupported
	defer func() {
		clipboard.Unsupported = ov
	}()
	clipboard.Unsupported = true

	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	color.NoColor = true
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	out.Stdout = stdoutBuf
	stdout = stdoutBuf
	out.Stderr = stderrBuf
	stderr = stderrBuf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		stderr = os.Stderr
		out.Stderr = os.Stderr
	}()

	// terminal=false
	ctx = ctxutil.WithTerminal(ctx, false)

	// gopass show foo
	// -> w/o terminal
	// -> Print password
	// for use in scripts
	t.Run("gopass show foo", func(t *testing.T) {
		// terminal=false
		ctx = ctxutil.WithTerminal(ctx, false)
		// initialize context with config values, also detects if we're running in a terminal
		ctx = act.Store.WithStoreConfig(ctx)

		c := gptest.CliCtx(ctx, t, "foo")
		require.NoError(t, act.Show(c))
		assert.NotContains(t, stderrBuf.String(), "WARNING")
		assert.Contains(t, stdoutBuf.String(), "secret")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})

	// gopass show -c foo
	// -> Copy to clipboard
	t.Run("gopass show -c foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.Contains(t, stderrBuf.String(), "WARNING")
		assert.NotContains(t, stdoutBuf.String(), "secret")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})

	// gopass show -C foo
	// -> Copy to clipboard AND print
	t.Run("gopass show -C foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"alsoclip": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.Contains(t, stderrBuf.String(), "WARNING")
		assert.Contains(t, stdoutBuf.String(), "secret")
		assert.Contains(t, stdoutBuf.String(), "second")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})

	// gopass show -f foo
	// -> ONLY print
	t.Run("gopass show -f foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"unsafe": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.NotContains(t, stderrBuf.String(), "WARNING")
		assert.Contains(t, stdoutBuf.String(), "secret")
		assert.Contains(t, stdoutBuf.String(), "second")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})

	// gopass show foo
	// -> Copy to clipboard
	t.Run("gopass show foo", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "foo")
		require.NoError(t, act.Show(c))
		assert.NotContains(t, stderrBuf.String(), "WARNING")
		assert.Contains(t, stdoutBuf.String(), "secret")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})

	// gopass show -c foo
	// -> Copy to clipboard
	t.Run("gopass show -c foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.Contains(t, stderrBuf.String(), "WARNING")
		assert.NotContains(t, stdoutBuf.String(), "secret")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})

	// gopass show -C foo
	// -> Copy to clipboard AND print
	t.Run("gopass show -C foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"alsoclip": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.Contains(t, stderrBuf.String(), "WARNING")
		assert.Contains(t, stdoutBuf.String(), "secret")
		assert.Contains(t, stdoutBuf.String(), "second")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})

	// gopass show -f foo
	// -> ONLY Print
	t.Run("gopass show -f foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"unsafe": "true"}, "foo")
		require.NoError(t, act.Show(c))
		assert.NotContains(t, stderrBuf.String(), "WARNING")
		assert.Contains(t, stdoutBuf.String(), "secret")
		assert.Contains(t, stdoutBuf.String(), "second")
		stdoutBuf.Reset()
		stderrBuf.Reset()
	})
}

func TestShowHandleRevision(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	t.Run("show foo", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtx(ctx, t)
		require.NoError(t, act.showHandleRevision(ctx, c, "foo", "HEAD"))
	})
}

func TestShowHandleError(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	// show foo
	c := gptest.CliCtx(ctx, t)
	require.Error(t, act.showHandleError(ctx, c, "foo", false, fmt.Errorf("test")))
	buf.Reset()
}

func TestShowPrintQR(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx) //nolint:ineffassign

	color.NoColor = true
	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	require.NoError(t, act.showPrintQR("foo", "bar"))
	buf.Reset()
}

func TestShowHasAliasDomain(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	sec := secrets.NewAKV()
	sec.SetPassword("foo")
	require.NoError(t, act.Store.Set(ctx, "websites/foo.de/user", sec))

	require.NoError(t, act.cfg.Set("", "domain-alias.foo.de.insteadOf", "foo.com"))

	alias := act.hasAliasDomain(ctx, "websites/foo.com/user")
	assert.Equal(t, "websites/foo.de/user", alias)
}
