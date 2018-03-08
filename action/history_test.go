package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestHistory(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = backend.WithSyncBackend(ctx, backend.GitCLI)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGMock)

	cfg := config.New()
	cfg.Root.Path = backend.FromPath(u.StoreDir(""))
	act, err := newAction(ctx, cfg, semver.Version{})
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// init git
	assert.NoError(t, act.gitInit(ctx, "", "foo bar", "foo.bar@example.org"))

	// insert bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Insert(ctx, c))

	// history bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.History(ctx, c))
}
