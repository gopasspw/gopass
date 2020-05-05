package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/muesli/goprogressbar"

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
	goprogressbar.Stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		goprogressbar.Stdout = os.Stdout
	}()
	color.NoColor = true

	// fsck
	assert.NoError(t, act.Fsck(clictx(ctx, t)))
	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Checking store integrity ...")
	assert.Contains(t, out, "[] Extra recipients on foo: [0xFEEDBEEF]")
	assert.Contains(t, out, "[] Pushed changes to git remote")
	buf.Reset()

	// fsck fo
	assert.NoError(t, act.Fsck(clictx(ctx, t, "fo")))
	out = strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Checking store integrity ...")
	assert.Contains(t, out, "[] Extra recipients on foo: [0xFEEDBEEF]")
	assert.Contains(t, out, "[] Pushed changes to git remote")

	buf.Reset()
}
