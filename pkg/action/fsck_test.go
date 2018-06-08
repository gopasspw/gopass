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
	"github.com/urfave/cli"
)

func TestFsck(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
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

	// fsck
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	assert.NoError(t, act.Fsck(ctx, c))
	assert.Equal(t, "Checking store integrity ...\n\n[] Extra recipients on foo: [0xFEEDBEEF]\n\n[] Pushed changes to git remote", strings.TrimSpace(buf.String()))
	buf.Reset()

	// fsck fo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"fo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Fsck(ctx, c))
	assert.Equal(t, "Checking store integrity ...\n\n[] Extra recipients on foo: [0xFEEDBEEF]\n\n[] Pushed changes to git remote", strings.TrimSpace(buf.String()))
	buf.Reset()
}
