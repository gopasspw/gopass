package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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

	// first add two entries
	var sec gopass.Secret
	sec = secrets.NewAKV()
	sec.SetPassword("123")
	require.NoError(t, sec.Set("bar", "zab"))
	require.NoError(t, act.Store.Set(ctx, "bar/baz", sec))
	buf.Reset()

	sec = secrets.NewAKV()
	sec.SetPassword("456")
	require.NoError(t, sec.Set("bar", "baz"))
	require.NoError(t, act.Store.Set(ctx, "bar/zab", sec))
	buf.Reset()

	require.NoError(t, act.Merge(gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "bar/baz", "bar/zab")))

	sec, err = act.Store.Get(ctx, "bar/baz")
	require.NoError(t, err)

	assert.Equal(t, "\n# Secret: bar/baz\n123\nbar: zab\n\n# Secret: bar/zab\n456\nbar: baz\n", string(sec.Bytes()))
}
