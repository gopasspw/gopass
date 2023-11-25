package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	aclip "github.com/atotto/clipboard"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	u := gptest.NewUnitTester(t)

	aclip.Unsupported = true

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	require.NoError(t, act.cfg.Set("", "core.notifications", "false"))
	require.NoError(t, act.cfg.Set("", "core.cliptimeout", "1"))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// create
	c := gptest.CliCtx(ctx, t)

	require.Error(t, act.Create(c))
	buf.Reset()
}
