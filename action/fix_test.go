package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestFix(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.NoError(t, act.Store.Set(ctx, "yaml/valid", secret.New("foo", "---\nbar: baz")))
	assert.NoError(t, act.Store.Set(ctx, "yaml/invalid1", secret.New("foo", "---\nbar")))
	assert.NoError(t, act.Store.Set(ctx, "yaml/invalid2", secret.New("foo", "bar:")))

	app := cli.NewApp()

	// fix
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Fix(ctx, c))

	// fix --force
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "force",
		Usage: "force",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--force"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Fix(ctx, c))

	// fix --check
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	sf = cli.BoolFlag{
		Name:  "check",
		Usage: "check",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--check=true"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Fix(ctx, c))
}
