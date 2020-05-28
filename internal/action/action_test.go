package action

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/gptest"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func newMock(ctx context.Context, u *gptest.Unit) (*Action, error) {
	cfg := config.Load()
	cfg.Path = u.StoreDir("")

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

	cfg.Path = filepath.Join(td, "store")
	assert.NoError(t, os.MkdirAll(cfg.Path, 0700))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(cfg.Path, plain.IDFile), []byte("foobar"), 0600))
	_, err = New(ctx, cfg, sv)
	assert.NoError(t, err)
}
