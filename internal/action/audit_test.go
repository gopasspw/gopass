package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAudit(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	t.Run("expect audit complaints on very weak passwords", func(t *testing.T) {
		sec := secrets.NewAKV()
		sec.SetPassword("123")
		require.NoError(t, act.Store.Set(ctx, "bar", sec))
		require.NoError(t, act.Store.Set(ctx, "baz", sec))

		require.Error(t, act.Audit(gptest.CliCtx(ctx, t)))
		buf.Reset()
	})

	t.Run("test with filter and very passwords", func(t *testing.T) {
		c := gptest.CliCtx(ctx, t, "foo")
		require.Error(t, act.Audit(c))
		buf.Reset()
	})

	t.Run("test empty store", func(t *testing.T) {
		for _, v := range []string{"foo", "bar", "baz"} {
			require.NoError(t, act.Store.Delete(ctx, v))
		}
		require.NoError(t, act.Audit(gptest.CliCtx(ctx, t)))
		assert.Contains(t, "No secrets found", buf.String())
		buf.Reset()
	})
}
