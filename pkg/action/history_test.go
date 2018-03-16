package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/config"
	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestHistory(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithDebug(ctx, true)
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)

	cfg := config.New()
	cfg.Root.Path = backend.FromPath(u.StoreDir(""))
	act, err := newAction(ctx, cfg, semver.Version{})
	assert.NoError(t, err)
	assert.NoError(t, act.Initialized(ctx, nil))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// init git
	assert.NoError(t, act.gitInit(ctx, "", "foo bar", "foo.bar@example.org"))
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
}
