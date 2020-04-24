package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestEdit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// edit
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	assert.Error(t, act.Edit(ctx, c))
	buf.Reset()

	// edit foo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Edit(ctx, c))
	buf.Reset()
}
