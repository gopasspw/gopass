package action

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	sec := secrets.New()
	require.NoError(t, sec.Set("username", "admin"))
	sec.SetPassword("hunter2")
	require.NoError(t, act.Store.Set(ctx, "server/local/mysql", sec))

	infile := filepath.Join(u.Dir, "my.cnf.tpl")
	err = os.WriteFile(infile, []byte(`[client]
host=127.0.0.1
port=3306
user={{ getval "server/local/mysql" "username" }}
password={{ getpw "server/local/mysql" }}`), 0o644)
	require.NoError(t, err)

	t.Run("process template", func(t *testing.T) {
		defer buf.Reset()

		err := act.Process(gptest.CliCtx(ctx, t, infile))
		require.NoError(t, err)
		assert.Equal(t, `[client]
host=127.0.0.1
port=3306
user=admin
password=hunter2
`, buf.String(), "processed template")
	})
}
