package action

import (
	"bytes"
	"context"
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
	"github.com/urfave/cli/v3"
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

		err := act.Process(ctx, gptest.CliCtx(ctx, t, infile))
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
		require.NoError(t, act.Process(ctx, c))
		assert.Contains(t, buf.String(), "password=hunter2")
	})

	t.Run("allow-path denies secret outside prefix", func(t *testing.T) {
		defer buf.Reset()

		// Template references server/local/mysql but only other/path is allowed.
		c := cliCtxWithAllowPaths(ctx, t, []string{"other/path"}, infile)
		err := act.Process(ctx, c)
		require.Error(t, err, "template must fail when secret is outside allowed paths")
	})
}

// cliCtxWithAllowPaths builds a *cli.Command that has the --allow-path
// StringSlice flag populated with the given values and the positional argument
// set to file.
func cliCtxWithAllowPaths(ctx context.Context, t *testing.T, allowPaths []string, file string) *cli.Command {
	t.Helper()

	allArgs := make([]string, 0, len(allowPaths)+2)
	allArgs = append(allArgs, "test")
	for _, p := range allowPaths {
		allArgs = append(allArgs, "--allow-path="+p)
	}
	allArgs = append(allArgs, file)

	var captured *cli.Command

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringSliceFlag{Name: "allow-path", Aliases: []string{"p"}},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			captured = cmd

			return nil
		},
	}

	require.NoError(t, cmd.Run(ctx, allArgs))

	if captured == nil {
		return cmd
	}

	return captured
}
