package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/urfave/cli.v1"
)

func TestHistory(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	cfg := config.New()
	cfg.Root.Path = backend.FromPath(u.StoreDir(""))
	act, err := newAction(ctx, cfg, semver.Version{})
	require.NoError(t, err)
	require.NotNil(t, act)
	require.NoError(t, act.Initialized(ctx, nil))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// init git
	require.NoError(t, act.rcsInit(ctx, "", "foo bar", "foo.bar@example.org"))
	buf.Reset()

	// insert bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Insert(ctx, c))
	buf.Reset()

	// history bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.History(ctx, c))
	buf.Reset()

	// history --password bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.StringFlag{
		Name:  "password",
		Usage: "password",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--password=true", "bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.History(ctx, c))
	buf.Reset()
}
