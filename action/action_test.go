package action

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver"
	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
)

func newMock(ctx context.Context, u *gptest.Unit) (*Action, error) {
	cfg := config.New()
	cfg.Root.Path = u.StoreDir("")

	sv := semver.Version{}
	gpg := gpgmock.New()

	return newAction(ctx, cfg, sv, gpg)
}

func TestAction(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)
	assert.Equal(t, "action.test", act.Name)

	assert.Contains(t, act.String(), u.StoreDir(""))
	assert.Equal(t, true, act.HasGPG())
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

	cfg.Root.Path = filepath.Join(td, "store")
	_, err = New(ctx, cfg, sv)
	assert.NoError(t, err)
}

func TestUmask(t *testing.T) {
	for _, vn := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		for in, out := range map[string]int{
			"002":      02,
			"0777":     0777,
			"000":      0,
			"07557575": 077,
		} {
			assert.NoError(t, os.Setenv(vn, in))
			assert.Equal(t, out, umask())
			assert.NoError(t, os.Unsetenv(vn))
		}
	}
}

func TestGpgOpts(t *testing.T) {
	for _, vn := range []string{"GOPASS_GPG_OPTS", "PASSWORD_STORE_GPG_OPTS"} {
		for in, out := range map[string][]string{
			"": nil,
			"--decrypt --armor --recipient 0xDEADBEEF": {"--decrypt", "--armor", "--recipient", "0xDEADBEEF"},
		} {
			assert.NoError(t, os.Setenv(vn, in))
			assert.Equal(t, out, gpgOpts())
			assert.NoError(t, os.Unsetenv(vn))
		}
	}
}
