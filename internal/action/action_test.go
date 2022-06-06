package action

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func newMock(ctx context.Context, path string) (*Action, error) {
	cfg := config.Load()
	cfg.Path = path

	if !backend.HasCryptoBackend(ctx) {
		ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	}
	ctx = backend.WithStorageBackend(ctx, backend.GitFS)
	act, err := newAction(cfg, semver.Version{}, false)
	if err != nil {
		return nil, err
	}

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(cli.NewApp(), fs, nil)
	c.Context = ctx
	if err := act.IsInitialized(c); err != nil {
		return nil, err
	}

	return act, nil
}

func TestAction(t *testing.T) {
	t.Parallel()

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
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
	t.Parallel()

	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	cfg := config.New()
	sv := semver.Version{}

	t.Run("init a new store", func(t *testing.T) { //nolint:paralleltest
		_, err = New(cfg, sv)
		require.NoError(t, err)
	})

	t.Run("init an existing plain store", func(t *testing.T) { //nolint:paralleltest
		cfg.Path = filepath.Join(td, "store")
		assert.NoError(t, os.MkdirAll(cfg.Path, 0o700))
		assert.NoError(t, os.WriteFile(filepath.Join(cfg.Path, plain.IDFile), []byte("foobar"), 0o600))
		_, err = New(cfg, sv)
		assert.NoError(t, err)
	})
}
