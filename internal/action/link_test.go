package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLink(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewNoWrites().WithConfig(context.Background())
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

	require.NoError(t, act.Link(gptest.CliCtx(ctx, t, "bar/baz", "other/linkdest")))

	// original secret should be equal to the linkdest
	oSec, err := act.Store.Get(ctx, "bar/baz")
	require.NoError(t, err)

	lSec, err := act.Store.Get(ctx, "other/linkdest")
	require.NoError(t, err)

	assert.Equal(t, oSec.Bytes(), lSec.Bytes())

	// update the original, linkdest should still be the same
	oSec.SetPassword("456")
	require.NoError(t, act.Store.Set(ctx, "bar/baz", oSec))
	buf.Reset()

	oSec, err = act.Store.Get(ctx, "bar/baz")
	require.NoError(t, err)

	lSec, err = act.Store.Get(ctx, "other/linkdest")
	require.NoError(t, err)

	assert.Equal(t, oSec.Bytes(), lSec.Bytes())
}
