package action

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	git "github.com/gopasspw/gopass/internal/backend/rcs/git/cli"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aGitRepo creates and initializes a small git repo
func aGitRepo(ctx context.Context, u *gptest.Unit, t *testing.T, name string) string {
	gd := filepath.Join(u.Dir, name)
	assert.NoError(t, os.MkdirAll(gd, 0700))

	_, err := git.Open(gd, "")
	assert.Error(t, err)

	idf := filepath.Join(gd, ".gpg-id")
	assert.NoError(t, ioutil.WriteFile(idf, []byte("0xDEADBEEF"), 0600))

	gr, err := git.Init(ctx, gd, "Nobody", "foo.bar@example.org")
	assert.NoError(t, err)
	assert.NotNil(t, gr)

	return gd
}

func TestClone(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithRCSBackend(ctx, backend.GitCLI)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	// no args
	c := gptest.CliCtx(ctx, t)
	assert.Error(t, act.Clone(c))

	// clone to initialized store
	assert.Error(t, act.clone(ctx, "/tmp/non-existing-repo.git", "", filepath.Join(u.Dir, "store")))

	// clone to mount
	gd := aGitRepo(ctx, u, t, "other-repo")
	assert.NoError(t, act.clone(ctx, gd, "gd", filepath.Join(u.Dir, "mount")))
}

func TestCloneBackendIsStoredForMount(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	cfg := config.Load()
	cfg.Path = u.StoreDir("")

	act, err := newAction(ctx, cfg, semver.Version{})
	require.NoError(t, err)
	require.NotNil(t, act)

	c := gptest.CliCtx(ctx, t)
	require.NoError(t, act.Initialized(c))

	repo := aGitRepo(ctx, u, t, "my-project")

	c = gptest.CliCtx(ctx, t, repo, "the-project")
	assert.NoError(t, act.Clone(c))

	require.NotNil(t, act.cfg.Mounts["the-project"])
}

func TestCloneGetGitConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	name, email, err := act.cloneGetGitConfig(ctx, "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "", name)
	assert.Equal(t, "", email)
}

func TestDetectCryptoBackend(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	gpgdir := filepath.Join(tempdir, ".password-store-gpg")
	gpgfn := filepath.Join(gpgdir, ".gpg-id")
	assert.NoError(t, os.Mkdir(gpgdir, 0755))
	assert.NoError(t, ioutil.WriteFile(gpgfn, []byte("foobar"), 0644))

	xcdir := filepath.Join(tempdir, ".password-store-xc")
	xcfn := filepath.Join(xcdir, ".xc-ids")
	assert.NoError(t, os.Mkdir(xcdir, 0755))
	assert.NoError(t, ioutil.WriteFile(xcfn, []byte("foobar"), 0644))

	assert.Equal(t, backend.GPGCLI, detectCryptoBackend(ctx, gpgdir))
	assert.Equal(t, backend.XC, detectCryptoBackend(ctx, xcdir))
}
