package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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
		assert.Equal(t, "\nfoobar", buf.String())
		buf.Reset()
	})

	t.Run("insert baz via stdin w/ newline", func(t *testing.T) {
		assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar\n"), false))
		buf.Reset()

		assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
		assert.Equal(t, "\nfoobar\n", buf.String())
		buf.Reset()
	})

	t.Run("insert zab#key", func(t *testing.T) {
		assert.NoError(t, act.insertYAML(ctx, "zab", "key", []byte("foobar"), nil))
		buf.Reset()
	})

	t.Run("insert --multiline bar baz", func(t *testing.T) {
		assert.NoError(t, act.Insert(gptest.CliCtxWithFlags(ctx, t, map[string]string{"multiline": "true"}, "bar", "baz")))
		buf.Reset()
	})
}

func TestInsertStdin(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	ctx = ctxutil.WithStdin(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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
