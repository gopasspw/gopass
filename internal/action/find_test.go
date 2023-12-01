package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestFind(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	require.NoError(t, act.cfg.Set("", "generate.autoclip", "false"))

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()
	color.NoColor = true

	actName := "action.test"
	if runtime.GOOS == "windows" {
		actName = "action.test.exe"
	}

	// find
	c := gptest.CliCtx(ctx, t)
	if err := act.FindFuzzy(c); err == nil || err.Error() != fmt.Sprintf("Usage: %s find <pattern>", actName) {
		t.Errorf("Should fail: %s", err)
	}

	// find fo (with fuzzy search)
	c = gptest.CliCtxWithFlags(ctx, t, nil, "fo")
	require.NoError(t, act.FindFuzzy(c))
	assert.Contains(t, strings.TrimSpace(buf.String()), "Found exact match in \"foo\"\nsecret")
	buf.Reset()

	// find fo (no fuzzy search)
	c = gptest.CliCtxWithFlags(ctx, t, nil, "fo")
	require.NoError(t, act.Find(c))
	assert.Equal(t, "foo", strings.TrimSpace(buf.String()))
	buf.Reset()

	// testing the safecontent case
	require.NoError(t, act.cfg.Set("", "show.safecontent", "true"))
	c.Context = ctx
	require.NoError(t, act.FindFuzzy(c))
	buf.Reset()

	// testing with the clip flag set
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"clip": "true"}, "fo")
	require.NoError(t, act.FindFuzzy(c))
	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in \"foo\"")
	buf.Reset()

	// safecontent case with force flag set
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"unsafe": "true"}, "fo")
	require.NoError(t, act.FindFuzzy(c))
	out = strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Found exact match in \"foo\"\nsecret")
	buf.Reset()

	// stopping with the safecontent tests
	require.NoError(t, act.cfg.Set("", "show.safecontent", "false"))

	// find yo
	c = gptest.CliCtx(ctx, t, "yo")
	require.Error(t, act.FindFuzzy(c))
	buf.Reset()

	// add some secrets
	sec := secrets.NewAKV()
	sec.SetPassword("foo")
	_, err = sec.Write([]byte("bar"))
	require.NoError(t, err)
	require.NoError(t, act.Store.Set(ctx, "bar/baz", sec))
	require.NoError(t, act.Store.Set(ctx, "bar/zab", sec))
	buf.Reset()

	// find bar
	c = gptest.CliCtx(ctx, t, "bar")
	require.NoError(t, act.FindFuzzy(c))
	assert.Equal(t, "bar/baz\nbar/zab", strings.TrimSpace(buf.String()))
	buf.Reset()

	// find w/o callback
	c = gptest.CliCtx(ctx, t)
	require.NoError(t, act.find(ctx, c, "foo", nil, false))
	assert.Equal(t, "foo", strings.TrimSpace(buf.String()))
	buf.Reset()

	// findSelection w/o callback
	c = gptest.CliCtx(ctx, t)
	require.Error(t, act.findSelection(ctx, c, []string{"foo", "bar"}, "fo", nil))

	// findSelection w/o options
	c = gptest.CliCtx(ctx, t)
	require.Error(t, act.findSelection(ctx, c, nil, "fo", func(_ context.Context, _ *cli.Context, _ string, _ bool) error { return nil }))
}
