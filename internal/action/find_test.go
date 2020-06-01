package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	actName := "action.test"
	if runtime.GOOS == "windows" {
		actName = "action.test.exe"
	}

	// find
	c := gptest.CliCtx(ctx, t)
	if err := act.Find(c); err == nil || err.Error() != fmt.Sprintf("Usage: %s find <NEEDLE>", actName) {
		t.Errorf("Should fail: %s", err)
	}

	// find fo
	c = gptest.CliCtxWithFlags(ctx, t, nil, "fo")
	assert.NoError(t, act.Find(c))
	assert.Contains(t, strings.TrimSpace(buf.String()), "Found exact match in 'foo'\nsecret")
	buf.Reset()

	// find fo (no fuzzy search)
	c = gptest.CliCtxWithFlags(ctx, t, nil, "fo")
	assert.NoError(t, act.FindNoFuzzy(c))
	assert.Equal(t, strings.TrimSpace(buf.String()), "foo")
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

	// find w/o callback
	c = gptest.CliCtx(ctx, t)
	assert.NoError(t, act.find(ctx, c, "foo", nil))
	assert.Equal(t, "foo", strings.TrimSpace(buf.String()))
	buf.Reset()

	// findSelection w/o callback
	c = gptest.CliCtx(ctx, t)
	assert.Error(t, act.findSelection(ctx, c, []string{"foo", "bar"}, "fo", nil))

	// findSelection w/o options
	c = gptest.CliCtx(ctx, t)
	assert.Error(t, act.findSelection(ctx, c, nil, "fo", func(_ context.Context, _ *cli.Context, _ string, _ bool) error { return nil }))
}
