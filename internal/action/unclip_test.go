package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/rcs"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
)

func TestUnclip(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	app := cli.NewApp()

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.IntFlag{
		Name:  "timeout",
		Usage: "timeout",
	}
	assert.NoError(t, sf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--timeout=0"}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.Error(t, act.Unclip(c))
}
