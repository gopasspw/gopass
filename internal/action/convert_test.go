package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithPasswordCallback(ctx, func(s string, b bool) ([]byte, error) {
		return []byte("foo"), nil
	})
	ctx = ctxutil.WithPasswordPurgeCallback(ctx, func(s string) {})

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

	require.NoError(t, act.Convert(gptest.CliCtxWithFlags(ctx, t, map[string]string{
		"move":    "true",
		"storage": "fs",
		"crypto":  "age",
	})))
	// TODO: validate converted store. t.Logf("Buffer: %s", buf.String()).
}
