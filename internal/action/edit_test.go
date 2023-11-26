package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestEdit(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// edit
	require.Error(t, act.Edit(gptest.CliCtx(ctx, t)))
	buf.Reset()

	// edit foo (existing)
	require.Error(t, act.Edit(gptest.CliCtx(ctx, t, "foo")))
	buf.Reset()

	// edit bar (new)
	require.Error(t, act.Edit(gptest.CliCtx(ctx, t, "foo")))
	buf.Reset()
}

func TestEditUpdate(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	content := []byte("foobar")
	// no changes
	require.NoError(t, act.editUpdate(ctx, "foo", content, content, false, "test"))
	buf.Reset()

	// changes
	nContent := []byte("barfoo")
	require.NoError(t, act.editUpdate(ctx, "foo", content, nContent, false, "test"))
	buf.Reset()
}
