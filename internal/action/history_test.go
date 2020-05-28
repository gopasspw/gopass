package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestHistory(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	cfg := config.New()
	cfg.Path = u.StoreDir("")
	act, err := newAction(ctx, cfg, semver.Version{})
	require.NoError(t, err)
	require.NotNil(t, act)
	require.NoError(t, act.Initialized(c))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	// init git
	require.NoError(t, act.rcsInit(ctx, "", "foo bar", "foo.bar@example.org"))
	buf.Reset()

	// insert bar
	assert.NoError(t, act.Insert(gptest.CliCtx(ctx, t, "bar")))
	buf.Reset()

	// history bar
	assert.NoError(t, act.History(gptest.CliCtx(ctx, t, "bar")))
	buf.Reset()

	// history --password bar
	assert.NoError(t, act.History(gptest.CliCtxWithFlags(ctx, t, map[string]string{"password": "true"}, "bar")))
	buf.Reset()
}
