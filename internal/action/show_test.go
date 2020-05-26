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
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// show foo
	c := gptest.CliCtx(ctx, t, "foo")
	assert.NoError(t, act.Show(c))
	assert.Contains(t, buf.String(), "secret")
	buf.Reset()

	// show --sync foo
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"sync": "true"}, "foo")
	assert.NoError(t, act.Show(c))
	assert.Contains(t, buf.String(), "secret")
	buf.Reset()

	// show dir

	// first add another entry in a subdir
	assert.NoError(t, act.Store.Set(ctx, "bar/baz", secret.New("123", "---\nbar: zab")))
	buf.Reset()

	c = gptest.CliCtx(ctx, t, "bar")
	assert.NoError(t, act.Show(c))
	assert.Equal(t, "bar\n└── baz\n\n", buf.String())
	buf.Reset()

	// show twoliner with safecontent enabled
	ctx = ctxutil.WithShowSafeContent(ctx, true)

	c = gptest.CliCtx(ctx, t, "bar/baz")
	assert.NoError(t, act.Show(c))
	assert.Equal(t, "---\nbar: zab", buf.String())
	buf.Reset()

	// show foo with safecontent enabled, should error out
	c = gptest.CliCtx(ctx, t, "foo")
	assert.NoError(t, act.Show(c))
	assert.NotContains(t, buf.String(), "secret")
	buf.Reset()

	// show foo with safecontent enabled, with the force flag
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "foo")
	assert.NoError(t, act.Show(c))
	assert.Contains(t, buf.String(), "secret")
	buf.Reset()

	// show twoliner with safecontent enabled, but with the clip flag, which should copy just the secret
	ctx = ctxutil.WithShowSafeContent(ctx, true)
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "bar/baz")

	assert.NoError(t, act.Show(c))
	assert.NotContains(t, buf.String(), "123")
	buf.Reset()
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

	// autoclip=true
	ctx = ctxutil.WithAutoClip(ctx, true)
	// terminal=false
	ctx = ctxutil.WithTerminal(ctx, false)

	// gopass show foo
	// -> with AutoClip
	// -> w/o terminal
	// -> Print password
	// for use in scripts
	t.Run("gopass show foo", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "foo")
		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		buf.Reset()
	})

	// autoclip=true
	ctx = ctxutil.WithAutoClip(ctx, true)
	// terminal=true
	ctx = ctxutil.WithTerminal(ctx, true)

	// gopass show foo
	// -> with AutoClip
	// -> with terminal
	// -> Copy to clipboard
	// for interactive use
	t.Run("gopass show foo", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "foo")
		assert.NoError(t, act.Show(c))
		assert.Contains(t, buf.String(), "WARNING")
		assert.NotContains(t, buf.String(), "secret")
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
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "foo")
		assert.NoError(t, act.Show(c))
		assert.NotContains(t, buf.String(), "WARNING")
		assert.Contains(t, buf.String(), "secret")
		assert.Contains(t, buf.String(), "second")
		buf.Reset()
	})

	// autoclip=false
	ctx = ctxutil.WithAutoClip(ctx, false)
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
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "foo")
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

	// show foo
	c := gptest.CliCtx(ctx, t)
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

	// show foo
	c := gptest.CliCtx(ctx, t)
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
