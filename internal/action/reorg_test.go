package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestReorg(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := t.Context()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithTerminal(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	t.Run("move foo to bar", func(t *testing.T) {
		defer buf.Reset()

		// create a secret
		sec := secrets.NewAKVWithData("foo", nil, "", false)
		require.NoError(t, act.Store.Set(ctxutil.WithGitCommit(ctx, false), "foo", sec))
		buf.Reset()

		initial := []string{"foo"}
		modified := []string{"bar"}

		require.NoError(t, act.ReorgAfterEdit(ctx, initial, modified))

		// check that foo is now bar
		_, err := act.Store.Get(ctx, "bar")
		require.NoError(t, err)

		// check that foo is gone
		_, err = act.Store.Get(ctx, "foo")
		require.Error(t, err)
	})
}
