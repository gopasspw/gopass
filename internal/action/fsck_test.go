package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/can"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()
	color.NoColor = true

	// generate foo/bar
	c := gptest.CliCtx(ctx, t, "foo/bar", "24")
	require.NoError(t, act.Generate(c), "gopass generate foo/bar 24")
	buf.Reset()

	// fsck
	require.NoError(t, act.Fsck(gptest.CliCtx(ctx, t)))
	output := strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	assert.Contains(t, output, "Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck (hidden)
	require.NoError(t, act.Fsck(gptest.CliCtx(ctxutil.WithHidden(ctx, true), t)))
	output = strings.TrimSpace(buf.String())
	assert.NotContains(t, output, "Checking password store integrity ...")
	assert.NotContains(t, output, "Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck --decrypt
	require.NoError(t, act.Fsck(gptest.CliCtxWithFlags(ctx, t, map[string]string{"decrypt": "true"})))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	assert.Contains(t, output, "Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck foo
	require.NoError(t, act.Fsck(gptest.CliCtx(ctx, t, "foo")))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	assert.Contains(t, output, "Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()
}

func TestFsckGpg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	u := gptest.NewGUnitTester(t)

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()
	color.NoColor = true

	sub, err := act.Store.GetSubStore("")
	require.NoError(t, err)
	require.NoError(t, sub.ImportMissingPublicKeys(ctx, can.KeyID()))

	// generate foo/bar
	c := gptest.CliCtx(ctx, t, "foo/bar", "24")
	require.NoError(t, act.Generate(c), "gopass generate foo/bar 24")
	buf.Reset()

	// fsck
	require.NoError(t, act.Fsck(gptest.CliCtx(ctx, t)))
	output := strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	buf.Reset()

	// fsck (hidden)
	require.NoError(t, act.Fsck(gptest.CliCtx(ctxutil.WithHidden(ctx, true), t)))
	output = strings.TrimSpace(buf.String())
	assert.NotContains(t, output, "Checking password store integrity ...")
	assert.NotContains(t, output, "Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck --decrypt
	require.NoError(t, act.Fsck(gptest.CliCtxWithFlags(ctx, t, map[string]string{"decrypt": "true"})))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	buf.Reset()

	// fsck foo
	require.NoError(t, act.Fsck(gptest.CliCtx(ctx, t, "foo")))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking password store integrity ...")
	buf.Reset()
}
