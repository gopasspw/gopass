package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestGrep(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	c := gptest.CliCtx(ctx, t, "foo")
	t.Run("empty store", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.Grep(c))
	})

	t.Run("add some secret", func(t *testing.T) {
		defer buf.Reset()
		sec := secrets.NewAKV()
		sec.SetPassword("foobar")
		_, err := sec.Write([]byte("foobar"))
		require.NoError(t, err)
		require.NoError(t, act.Store.Set(ctx, "foo", sec))
	})

	t.Run("should find existing", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.Grep(c))
	})

	t.Run("RE2", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"regexp": "true"}, "f..bar")
		require.NoError(t, act.Grep(c))
	})
}
