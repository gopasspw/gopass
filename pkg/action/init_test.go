package action

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestInit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithDebug(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo.bar@example.org"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Initialized(ctx, c))
	assert.Error(t, act.Init(ctx, c))
	assert.Error(t, act.InitOnboarding(ctx, c))
	crypto := act.Store.Crypto(ctx, "")
	assert.Equal(t, true, act.initHasUseablePrivateKeys(ctx, crypto, ""))
	assert.Error(t, act.initCreatePrivateKey(ctx, crypto, "", "foo bar", "foo.bar@example.org"))
	buf.Reset()

	// un-initialize the store
	assert.NoError(t, os.Remove(filepath.Join(u.StoreDir(""), ".gpg-id")))
	assert.Error(t, act.Initialized(ctx, c))
	buf.Reset()
}

func TestInitParseContext(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	for _, tc := range []struct {
		name  string
		args  []string
		check func(context.Context) error
	}{
		{
			name: "crypto xc",
			args: []string{"--crypto=xc"},
			check: func(ctx context.Context) error {
				if backend.GetCryptoBackend(ctx) != backend.XC {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
		{
			name: "rcs noop",
			args: []string{"--rcs=noop"},
			check: func(ctx context.Context) error {
				if backend.GetRCSBackend(ctx) != backend.Noop {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
		{
			name: "nogit",
			args: []string{"--nogit"},
			check: func(ctx context.Context) error {
				if backend.GetRCSBackend(ctx) != backend.Noop {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
		{
			name: "default",
			args: []string{},
			check: func(ctx context.Context) error {
				if backend.GetRCSBackend(ctx) != backend.GitCLI {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
	} {
		app := cli.NewApp()
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		sf := cli.StringFlag{
			Name:  "crypto",
			Usage: "crypto",
		}
		assert.NoError(t, sf.ApplyWithError(fs))
		sf = cli.StringFlag{
			Name:  "rcs",
			Usage: "rcs",
		}
		assert.NoError(t, sf.ApplyWithError(fs))
		bf := cli.BoolFlag{
			Name:  "nogit",
			Usage: "nogit",
		}
		assert.NoError(t, bf.ApplyWithError(fs))
		assert.NoError(t, fs.Parse(tc.args), tc.name)
		c := cli.NewContext(app, fs, nil)
		assert.NoError(t, tc.check(initParseContext(context.Background(), c)), tc.name)
		buf.Reset()
	}
}
