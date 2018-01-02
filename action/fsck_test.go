package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestFsck(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	for _, tc := range []struct {
		name string
		args map[string]string
	}{
		{
			name: "fsck",
		},
		{
			name: "fsck --check",
			args: map[string]string{
				"check": "true",
			},
		},
		{
			name: "fsck --force",
			args: map[string]string{
				"force": "true",
			},
		},
		{
			name: "fsck --check --force",
			args: map[string]string{
				"check": "true",
				"force": "true",
			},
		},
	} {
		// fsck
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		args := make([]string, 0, len(tc.args)*2)
		for an, av := range tc.args {
			f := cli.BoolFlag{
				Name:  an,
				Usage: an,
			}
			assert.NoError(t, f.ApplyWithError(fs))
			args = append(args, "--"+an, av)
		}
		assert.NoError(t, fs.Parse(args))
		c := cli.NewContext(app, fs, nil)
		assert.NoError(t, act.Fsck(ctx, c))
	}
}
