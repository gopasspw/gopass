package action

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func newMock(ctx context.Context, u *gptest.Unit) (*Action, error) {
	cfg := config.Load()
	cfg.Root.Path = backend.FromPath(u.StoreDir(""))

	ctx = backend.WithRCSBackend(ctx, backend.Noop)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.FS)
	act, err := newAction(ctx, cfg, semver.Version{})
	if err != nil {
		return nil, err
	}
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(cli.NewApp(), fs, nil)
	c.Context = ctx
	if err := act.Initialized(c); err != nil {
		return nil, err
	}
	return act, nil
}

func TestAction(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	actName := "action.test"

	if runtime.GOOS == "windows" {
		actName = "action.test.exe"
	}

	assert.Equal(t, actName, act.Name)

	assert.Contains(t, act.String(), u.StoreDir(""))
	assert.Equal(t, 0, len(act.Store.Mounts()))
}

func TestNew(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	cfg := config.New()
	sv := semver.Version{}

	_, err = New(ctx, cfg, sv)
	require.NoError(t, err)

	cfg.Root.Path = backend.FromPath(filepath.Join(td, "store"))
	cfg.Root.Path.Crypto = backend.Plain
	cfg.Root.Path.RCS = backend.Noop
	_, err = New(ctx, cfg, sv)
	assert.NoError(t, err)
}

func clictx(ctx context.Context, t *testing.T, args ...string) *cli.Context {
	return clictxf(ctx, t, nil, args...)
}

func clictxf(ctx context.Context, t *testing.T, flags map[string]string, args ...string) *cli.Context {
	app := cli.NewApp()

	fs := flagset(t, flags, args)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	return c
}

func flagset(t *testing.T, flags map[string]string, args []string) *flag.FlagSet {
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	for k, v := range flags {
		if v == "true" || v == "false" {
			f := cli.BoolFlag{
				Name:  k,
				Usage: k,
			}
			assert.NoError(t, f.Apply(fs))
		} else {
			f := cli.StringFlag{
				Name:  k,
				Usage: k,
			}
			assert.NoError(t, f.Apply(fs))
		}
	}
	argl := []string{}
	for k, v := range flags {
		argl = append(argl, "--"+k+"="+v)
	}
	argl = append(argl, args...)
	assert.NoError(t, fs.Parse(argl))

	return fs
}
