package action

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/backend"
	git "github.com/gopasspw/gopass/internal/backend/storage/gitfs"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aGitRepo creates and initializes a small git repo.
func aGitRepo(ctx context.Context, u *gptest.Unit, t *testing.T, name string) string {
	gd := filepath.Join(u.Dir, name)
	assert.NoError(t, os.MkdirAll(gd, 0o700))

	_, err := git.New(gd)
	assert.Error(t, err)

	idf := filepath.Join(gd, ".gpg-id")
	assert.NoError(t, os.WriteFile(idf, []byte("0xDEADBEEF"), 0o600))

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
	ctx = backend.WithStorageBackend(ctx, backend.GitFS)

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

	t.Run("no args", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtx(ctx, t)
		assert.Error(t, act.Clone(c))
	})

	t.Run("clone to initialized store", func(t *testing.T) {
		defer buf.Reset()
		assert.Error(t, act.clone(ctx, "/tmp/non-existing-repo.git", "", filepath.Join(u.Dir, "store")))
	})

	t.Run("clone to mount", func(t *testing.T) {
		defer buf.Reset()
		gd := aGitRepo(ctx, u, t, "other-repo")
		assert.NoError(t, act.clone(ctx, gd, "gd", filepath.Join(u.Dir, "mount")))
	})
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

	act, err := newAction(cfg, semver.Version{}, false)
	require.NoError(t, err)
	require.NotNil(t, act)

	c := gptest.CliCtx(ctx, t)
	require.NoError(t, act.IsInitialized(c))

	repo := aGitRepo(ctx, u, t, "my-project")

	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"check-keys": "false"}, repo, "the-project")
	assert.NoError(t, act.Clone(c))

	require.NotNil(t, act.cfg.Mounts["the-project"])
}

func TestCloneGetGitConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	r1 := gptest.UnsetVars(termio.NameVars...)
	defer r1()
	r2 := gptest.UnsetVars(termio.EmailVars...)
	defer r2()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	name, email, err := act.cloneGetGitConfig(ctx, "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "0xDEADBEEF", name)
	assert.Equal(t, "0xDEADBEEF", email)
}
