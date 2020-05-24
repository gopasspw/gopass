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

	// insert bar
	assert.NoError(t, act.Insert(gptest.CliCtx(ctx, t, "bar")))

	// insert bar baz
	assert.NoError(t, act.Insert(gptest.CliCtx(ctx, t, "bar", "baz")))

	// insert baz via stdin
	assert.NoError(t, act.insertStdin(ctx, "baz", []byte("foobar"), false))
	buf.Reset()

	assert.NoError(t, act.show(ctx, gptest.CliCtx(ctx, t), "baz", false))
	assert.Equal(t, "foobar", buf.String())
	buf.Reset()

	// insert zab#key
	assert.NoError(t, act.insertYAML(ctx, "zab", "key", []byte("foobar"), nil))

	// insert --multiline bar baz
	assert.NoError(t, act.Insert(gptest.CliCtxWithFlags(ctx, t, map[string]string{"multiline": "true"}, "bar", "baz")))
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
