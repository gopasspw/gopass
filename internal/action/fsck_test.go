package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

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

	// fsck
	assert.NoError(t, act.Fsck(gptest.CliCtx(ctx, t)))
	output := strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking store integrity ...")
	assert.Contains(t, output, "[] Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck (hidden)
	assert.NoError(t, act.Fsck(gptest.CliCtx(ctxutil.WithHidden(ctx, true), t)))
	output = strings.TrimSpace(buf.String())
	assert.NotContains(t, output, "Checking store integrity ...")
	assert.NotContains(t, output, "[] Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck --decrypt
	assert.NoError(t, act.Fsck(gptest.CliCtxWithFlags(ctx, t, map[string]string{"decrypt": "true"})))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking store integrity ...")
	assert.Contains(t, output, "[] Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()

	// fsck fo
	assert.NoError(t, act.Fsck(gptest.CliCtx(ctx, t, "fo")))
	output = strings.TrimSpace(buf.String())
	assert.Contains(t, output, "Checking store integrity ...")
	assert.Contains(t, output, "[] Extra recipients on foo: [0xFEEDBEEF]")
	buf.Reset()
}
