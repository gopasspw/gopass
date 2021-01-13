package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/atotto/clipboard"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowMulti(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
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

	t.Run("show foo", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtx(ctx, t, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "secret")
	})

	t.Run("show --sync foo", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"sync": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "secret")
	})

	t.Run("show dir", func(t *testing.T) {
		// first add another entry in a subdir
		sec := secrets.NewKV()
		sec.SetPassword("123")
		sec.Set("bar", "zab")
		assert.NoError(t, act.Store.Set(ctx, "bar/baz", sec))
		buf.Reset()

		c := gptest.CliCtx(ctx, t, "bar")
		assert.NoError(t, act.Show(c))
		assert.Equal(t, "bar/\n└── baz\n\n", buf.String())
		buf.Reset()
	})

	t.Run("show twoliner with safecontent enabled", func(t *testing.T) {
		ctx = ctxutil.WithShowSafeContent(ctx, true)
		c := gptest.CliCtx(ctx, t, "bar/baz")

		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "bar: zab")
		assert.Contains(t, buf.String(), "password: ***")
		buf.Reset()
	})

	t.Run("show foo with safecontent enabled, should error out", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "foo")
		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "secret")
		buf.Reset()
	})

	t.Run("show foo with safecontent enabled, with the force flag", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"unsafe": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "secret")
		buf.Reset()
	})

	t.Run("show twoliner with safecontent enabled, but with the clip flag, which should copy just the secret", func(t *testing.T) {
		ctx = ctxutil.WithShowSafeContent(ctx, true)
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "bar/baz")

		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "123")
		buf.Reset()
	})

	t.Run("show entry with unsafe keys", func(t *testing.T) {
		sec := secrets.NewKV()
		sec.SetPassword("123")
		sec.Set("bar", "zab")
		sec.Set("foo", "baz")
		sec.Set("hello", "world")
		sec.Set("unsafe-keys", "foo, bar")
		assert.NoError(t, act.Store.Set(ctx, "unsafe/keys", sec))
		buf.Reset()

		ctx = ctxutil.WithShowSafeContent(ctx, true)
		c := gptest.CliCtx(ctx, t, "unsafe/keys")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "*****")
		assert.NotContains(t, buf.String(), "zab")
		assert.NotContains(t, buf.String(), "baz")
		buf.Reset()
	})

	t.Run("show twoliner with safecontent enabled", func(t *testing.T) {
		ctx = ctxutil.WithShowSafeContent(ctx, true)
		c := gptest.CliCtx(ctx, t, "bar/baz")

		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "bar: zab")
		assert.Contains(t, buf.String(), "password: ***")
		buf.Reset()
	})

	t.Run("show twoliner with parsing disabled and safecontent enabled", func(t *testing.T) {
		ctx = ctxutil.WithShowSafeContent(ctx, true)
		ctx = ctxutil.WithShowParsing(ctx, false)
		c := gptest.CliCtx(ctx, t, "bar/baz")

		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "bar: zab")
		// password should not show up neither be obstructed
		assert.NotContains(t, buf.String(), "123")
		assert.NotContains(t, buf.String(), "***")
		buf.Reset()
	})

	t.Run("show key with parsing enabled", func(t *testing.T) {
		ctx = ctxutil.WithShowParsing(ctx, true)
		c := gptest.CliCtx(ctx, t, "bar/baz", "bar")

		assert.NoError(t, act.Show(c))
		assert.Equal(t, "zab", buf.String())
		buf.Reset()
	})

	t.Run("show key with parsing disabled", func(t *testing.T) {
		ctx = ctxutil.WithShowParsing(ctx, false)
		c := gptest.CliCtx(ctx, t, "bar/baz", "bar")

		assert.NoError(t, act.Show(c))
		assert.Equal(t, "bar: zab\n", buf.String())
		buf.Reset()
	})

	t.Run("show nonexisting key with parsing enabled", func(t *testing.T) {
		ctx = ctxutil.WithShowParsing(ctx, true)
		c := gptest.CliCtx(ctx, t, "bar/baz", "nonexisting")

		assert.Error(t, act.Show(c))
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
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
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
		ctx = act.Store.WithContext(ctx)

		c := gptest.CliCtx(ctx, t, "foo")
		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		buf.Reset()
	})

	// gopass show -c foo
	// -> Copy to clipboard
	t.Run("gopass show -c foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "WARNING")
		assert.NotContains(t, buf.String(), "secret")
		buf.Reset()
	})

	// gopass show -C foo
	// -> Copy to clipboard AND print
	t.Run("gopass show -C foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"alsoclip": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		assert.Contains(t, buf.String(), "second")
		buf.Reset()
	})

	// gopass show -f foo
	// -> ONLY print
	t.Run("gopass show -f foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"unsafe": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		assert.Contains(t, buf.String(), "second")
		buf.Reset()
	})

	// gopass show foo
	// -> Copy to clipboard
	t.Run("gopass show foo", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "foo")
		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		buf.Reset()
	})

	// gopass show -c foo
	// -> Copy to clipboard
	t.Run("gopass show -c foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "WARNING")
		assert.NotContains(t, buf.String(), "secret")
		buf.Reset()
	})

	// gopass show -C foo
	// -> Copy to clipboard AND print
	t.Run("gopass show -C foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"alsoclip": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		assert.Contains(t, buf.String(), "second")
		buf.Reset()
	})

	// gopass show -f foo
	// -> ONLY Print
	t.Run("gopass show -f foo", func(t *testing.T) {
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"unsafe": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		assert.Contains(t, buf.String(), "second")
		buf.Reset()
	})
}

func TestShowHandleRevision(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
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

	t.Run("show foo", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtx(ctx, t)
		assert.NoError(t, act.showHandleRevision(ctx, c, "foo", "HEAD"))
	})
}

func TestShowHandleError(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
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

	// show foo
	c := gptest.CliCtx(ctx, t)
	assert.Error(t, act.showHandleError(ctx, c, "foo", false, fmt.Errorf("test")))
	buf.Reset()
}

func TestShowPrintQR(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
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

	assert.NoError(t, act.showPrintQR("foo", "bar"))
	buf.Reset()
}
