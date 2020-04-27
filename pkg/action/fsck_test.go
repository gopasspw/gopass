package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
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

	app := cli.NewApp()

	// fsck
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	c.Context = ctx
	assert.NoError(t, act.Fsck(c))
	out := strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Checking store integrity ...")
	assert.Contains(t, out, "[] Extra recipients on foo: [0xFEEDBEEF]")
	assert.Contains(t, out, "[] Pushed changes to git remote")
	buf.Reset()

	// fsck fo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"fo"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.NoError(t, act.Fsck(c))
	out = strings.TrimSpace(buf.String())
	assert.Contains(t, out, "Checking store integrity ...")
	assert.Contains(t, out, "[] Extra recipients on foo: [0xFEEDBEEF]")
	assert.Contains(t, out, "[] Pushed changes to git remote")

	buf.Reset()
}
