package action

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
)

func newMock(ctx context.Context, u *gptest.Unit) (*Action, error) {
	cfg := config.New()
	cfg.Root.Path = backend.FromPath(u.StoreDir(""))

	ctx = backend.WithSyncBackendString(ctx, "gitmock")
	ctx = backend.WithCryptoBackendString(ctx, "gpgmock")
	return newAction(ctx, cfg, semver.Version{})
}

func TestAction(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)
	assert.Equal(t, "action.test", act.Name)

	assert.Contains(t, act.String(), u.StoreDir(""))
	assert.Equal(t, 0, len(act.Store.Mounts()))
}

func TestNew(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	cfg := config.New()
	sv := semver.Version{}

	_, err = New(ctx, cfg, sv)
	assert.Error(t, err)

	cfg.Root.Path = backend.FromPath(filepath.Join(td, "store"))
	_, err = New(ctx, cfg, sv)
	assert.NoError(t, err)
}
