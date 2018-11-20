package action

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	if err := act.Initialized(ctx, nil); err != nil {
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

	assert.Equal(t, "action.test", act.Name)

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
