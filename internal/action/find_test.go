package action

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestFind(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAutoClip(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	app := cli.NewApp()

	actName := "action.test"

	if runtime.GOOS == "windows" {
		actName = "action.test.exe"
	}

	// find
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	c.Context = ctx
	if err := act.Find(c); err == nil || err.Error() != fmt.Sprintf("Usage: %s find <NEEDLE>", actName) {
		t.Errorf("Should fail: %s", err)
	}

	// find fo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"fo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Find(c))
	assert.Contains(t, strings.TrimSpace(buf.String()), "Found exact match in 'foo'\nsecret")
	buf.Reset()

	// testing the safecontent case
	ctx = ctxutil.WithShowSafeContent(ctx, true)
	c.Context = ctx
	assert.NoError(t, act.Find(c))
	buf.Reset()

	// testing with the clip flag set
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "fo")
	assert.NoError(t, act.Find(c))
	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in 'foo'")
	buf.Reset()

	// safecontent case with force flag set
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"force": "true"}, "fo")
	assert.NoError(t, act.Find(c))
	out = strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in 'foo'\nsecret")
	buf.Reset()

	// stopping with the safecontent tests
	ctx = ctxutil.WithShowSafeContent(ctx, false)

	// find yo
	c = gptest.CliCtx(ctx, t, "yo")
	assert.Error(t, act.Find(c))
	buf.Reset()

	// add some secrets
	assert.NoError(t, act.Store.Set(ctx, "bar/baz", secret.New("foo", "bar")))
	assert.NoError(t, act.Store.Set(ctx, "bar/zab", secret.New("foo", "bar")))
	buf.Reset()

	// find bar
	c = gptest.CliCtx(ctx, t, "bar")
	assert.NoError(t, act.Find(c))
	assert.Equal(t, "bar/baz\nbar/zab", strings.TrimSpace(buf.String()))
	buf.Reset()
}
