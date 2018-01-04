package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestFind(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	app := cli.NewApp()

	// find
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	if err := act.Find(ctx, c); err == nil || err.Error() != "Usage: action.test find arg" {
		t.Errorf("Should fail")
	}

	// find fo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"fo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Find(ctx, c))
	assert.Equal(t, "Found exact match in 'foo'\n0xDEADBEEF", strings.TrimSpace(buf.String()))
	buf.Reset()

	// find yo
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"yo"}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Find(ctx, c))
	buf.Reset()

	// add some secrets
	assert.NoError(t, act.Store.Set(ctx, "bar/baz", secret.New("foo", "bar")))
	assert.NoError(t, act.Store.Set(ctx, "bar/zab", secret.New("foo", "bar")))

	// find bar
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Find(ctx, c))
	assert.Equal(t, "bar/baz\nbar/zab", strings.TrimSpace(buf.String()))
	buf.Reset()
}
