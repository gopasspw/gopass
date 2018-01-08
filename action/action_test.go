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
	"github.com/stretchr/testify/assert"
)

func newStore(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, ".gpg-id"), []byte("0xDEADBEEF"), 0600); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dir, "foo.gpg"), []byte("0xDEADBEEF"), 0600)
}

func newMock(ctx context.Context, dir string) (*Action, error) {
	cfg := config.New()
	cfg.Root.Path = filepath.Join(dir, "store")
	sv := semver.Version{}
	gpg := gpgmock.New()

	if err := os.Setenv("GOPASS_CONFIG", filepath.Join(dir, ".gopass.yml")); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_HOMEDIR", dir); err != nil {
		return nil, err
	}
	if err := os.Unsetenv("PAGER"); err != nil {
		return nil, err
	}
	if err := os.Setenv("CHECKPOINT_DISABLE", "true"); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_NO_NOTIFY", "true"); err != nil {
		return nil, err
	}
	if err := newStore(cfg.Root.Path); err != nil {
		return nil, err
	}

	return newAction(ctx, cfg, sv, gpg)
}

func TestAction(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)
	assert.Equal(t, "action.test", act.Name)

	assert.Contains(t, act.String(), filepath.Join(td, "store"))
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
