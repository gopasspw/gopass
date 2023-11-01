package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	require.NoError(t, act.cfg.Set("", "generate.autoclip", "true"))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
	}()

	t.Run("insert bar", func(t *testing.T) {
		assert.NoError(t, act.Insert(gptest.CliCtx(ctx, t, "bar")))
		buf.Reset()
	})

	t.Run("insert bar baz", func(t *testing.T) {
		assert.NoError(t, act.Insert(gptest.CliCtx(ctx, t, "bar", "baz")))
		buf.Reset()
	})

	t.Run("insert baz via stdin w/o newline", func(t *testing.T) {
		assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar"), false))
		buf.Reset()

		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
		assert.Equal(t, "foobar\n", buf.String())
		buf.Reset()
	})

	t.Run("insert baz via stdin w/ newline", func(t *testing.T) {
		assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar\n"), false))
		buf.Reset()

		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
		assert.Equal(t, "foobar\n", buf.String())
		buf.Reset()
	})

	t.Run("insert baz via stdin w/ yaml", func(t *testing.T) {
		assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar\n---\nuser: name\nother: meh"), false))
		buf.Reset()

		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
		assert.Equal(t, "foobar\n---\nother: meh\nuser: name\n", buf.String())
		buf.Reset()
	})

	t.Run("insert baz via stdin w/ k-v", func(t *testing.T) {
		in := "foobar\ninvalid key-value\nOther: meh\nUser: name\nbody text\n"
		assert.NoError(t, act.insertStdin(ctx, "baz", []byte(in), false))
		buf.Reset()

		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
		assert.Equal(t, in, buf.String())
		buf.Reset()
	})

	t.Run("insert zab#key", func(t *testing.T) {
		ctx = ctxutil.WithInteractive(ctx, false)
		require.NoError(t, act.cfg.Set("", "core.showsafecontent", "true"))
		assert.NoError(t, act.insertYAML(ctx, "zabkey", "key", []byte("foobar"), nil))
		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "zabkey", false))
		assert.Contains(t, buf.String(), "key: foobar")
		buf.Reset()
	})

	t.Run("insert --multiline bar baz", func(t *testing.T) {
		assert.NoError(t, act.Insert(gptest.CliCtxWithFlags(ctx, t, map[string]string{"multiline": "true"}, "bar", "baz")))
		buf.Reset()
	})

	t.Run("insert key:value", func(t *testing.T) {
		assert.NoError(t, act.Insert(gptest.CliCtxWithFlags(ctx, t, nil, "keyvaltest", "baz:val")))
		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "keyvaltest", false))
		assert.Contains(t, buf.String(), "baz: val")
		buf.Reset()
	})

	t.Run("insert baz via stdin w/ yaml and input parsing and safecontent", func(t *testing.T) {
		assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar\n---\nuser: name\nother: 0123"), false))
		buf.Reset()

		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
		assert.Equal(t, "other: 83\nuser: name", buf.String())
		buf.Reset()
	})

	t.Run("insert baz via stdin w/ yaml", func(t *testing.T) {
		require.NoError(t, act.cfg.Set("", "core.showsafecontent", "false"))
		assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar\n---\nuser: name\nother: 0123"), false))
		buf.Reset()

		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
		assert.Equal(t, "foobar\n---\nother: 83\nuser: name\n", buf.String())
		buf.Reset()
	})
}

func TestInsertStdin(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithStdin(ctx, true)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	require.NoError(t, act.cfg.Set("", "generate.autoclip", "false"))

	buf := &bytes.Buffer{}
	ibuf := &bytes.Buffer{}
	out.Stdout = buf
	stdin = ibuf
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
		stdin = os.Stdin
	}()

	ibuf.WriteString("foobar")
	assert.Error(t, act.insert(ctx, gptest.CliCtx(ctx, t), "foo", "", false, false, false, false, nil))
	ibuf.Reset()
	buf.Reset()

	// force
	ibuf.WriteString("foobar")
	assert.NoError(t, act.insert(ctx, gptest.CliCtx(ctx, t), "foo", "", false, false, true, false, nil))
	ibuf.Reset()
	buf.Reset()

	// append
	ibuf.WriteString("foobar")
	assert.NoError(t, act.insert(ctx, gptest.CliCtx(ctx, t), "foo", "", false, false, false, true, nil))
	ibuf.Reset()
	buf.Reset()

	// echo
	ibuf.WriteString("foobar")
	assert.NoError(t, act.insert(ctx, gptest.CliCtx(ctx, t), "bar", "", true, false, false, false, nil))
	ibuf.Reset()
	buf.Reset()

	// multiline
	ibuf.WriteString("foobar")
	assert.NoError(t, act.insert(ctx, gptest.CliCtx(ctx, t), "baz", "", false, true, false, false, nil))
	ibuf.Reset()
	buf.Reset()
}
