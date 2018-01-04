package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestGit(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithDebug(ctx, true)
	ctx = ctxutil.WithVerbose(ctx, true)

	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// git init
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.GitInit(ctx, c))
	t.Logf("Out: %s", buf.String())
	buf.Reset()

	// git status
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"status"}))
	c = cli.NewContext(app, fs, nil)
	out := capture(t, func() error {
		return act.Git(ctx, c)
	})
	want := `On branch master
nothing to commit`
	if !strings.HasPrefix(out, want) {
		t.Errorf("'%s' != '%s'", want, out)
	}
	buf.Reset()
}
