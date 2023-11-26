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

func TestSync(t *testing.T) {
	u := gptest.NewUnitTester(t)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	t.Run("default", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.Sync(gptest.CliCtx(ctx, t)))
	})

	t.Run("sync --store=root", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.Sync(gptest.CliCtxWithFlags(ctx, t, map[string]string{"store": "root"})))
	})
}
