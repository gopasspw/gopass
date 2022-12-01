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

func TestAudit(t *testing.T) { //nolint:paralleltest
	u := gptest.NewUnitTester(t)
	defer u.Remove()

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

	t.Run("expect audit complaints on very weak passwords", func(t *testing.T) { //nolint:paralleltest
		sec := &secrets.Plain{}
		sec.SetPassword("123")
		assert.NoError(t, act.Store.Set(ctx, "bar", sec))
		assert.NoError(t, act.Store.Set(ctx, "baz", sec))

		assert.Error(t, act.Audit(gptest.CliCtx(ctx, t)))
		buf.Reset()
	})

	t.Run("test with filter and very passwords", func(t *testing.T) { //nolint:paralleltest
		c := gptest.CliCtx(ctx, t, "foo")
		assert.Error(t, act.Audit(c))
		buf.Reset()
	})

	t.Run("test empty store", func(t *testing.T) { //nolint:paralleltest
		for _, v := range []string{"foo", "bar", "baz"} {
			assert.NoError(t, act.Store.Delete(ctx, v))
		}
		assert.NoError(t, act.Audit(gptest.CliCtx(ctx, t)))
		assert.Contains(t, "No secrets found", buf.String())
		buf.Reset()
	})
}
