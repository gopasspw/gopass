package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestProcess(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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

	t.Run("allow-path permits matching prefix", func(t *testing.T) {
		defer buf.Reset()

		c := cliCtxWithAllowPaths(ctx, t, []string{"server/local"}, infile)
		require.NoError(t, act.Process(c))
		assert.Contains(t, buf.String(), "password=hunter2")
	})

	t.Run("allow-path denies secret outside prefix", func(t *testing.T) {
		defer buf.Reset()

		// Template references server/local/mysql but only other/path is allowed.
		c := cliCtxWithAllowPaths(ctx, t, []string{"other/path"}, infile)
		err := act.Process(c)
		require.Error(t, err, "template must fail when secret is outside allowed paths")
	})
}

// cliCtxWithAllowPaths builds a *cli.Context that has the --allow-path
// StringSlice flag populated with the given values and the positional argument
// set to file.
func cliCtxWithAllowPaths(ctx context.Context, t *testing.T, allowPaths []string, file string) *cli.Context {
	t.Helper()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)

	f := &cli.StringSliceFlag{Name: "allow-path", Aliases: []string{"p"}}
	require.NoError(t, f.Apply(fs))

	args := make([]string, 0, len(allowPaths)+1)
	for _, p := range allowPaths {
		args = append(args, "--allow-path="+p)
	}
	args = append(args, file)
	require.NoError(t, fs.Parse(args))

	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	return c
}
