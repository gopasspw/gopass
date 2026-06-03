package action

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
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
func aGitRepo(ctx context.Context, t *testing.T, u *gptest.Unit, name string) string {
	t.Helper()

	gd := filepath.Join(u.Dir, name)
	require.NoError(t, os.MkdirAll(gd, 0o700))

	_, err := git.New(gd)
	require.Error(t, err)

	idf := filepath.Join(gd, ".gpg-id")
	require.NoError(t, os.WriteFile(idf, []byte("0xDEADBEEF"), 0o600))

	gr, err := git.Init(ctx, gd, "Nobody", "foo.bar@example.org")
	require.NoError(t, err)
	assert.NotNil(t, gr)

	return gd
}

func TestClone(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithStorageBackend(ctx, backend.GitFS)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

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
		require.Error(t, act.Clone(ctx, c))
	})

	t.Run("clone to initialized store", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.clone(ctx, "/tmp/non-existing-repo.git", "", filepath.Join(u.Dir, "store")))
	})

	t.Run("clone to mount", func(t *testing.T) {
		defer buf.Reset()
		gd := aGitRepo(ctx, t, u, "other-repo")
		require.NoError(t, act.clone(ctx, gd, "gd", filepath.Join(u.Dir, "mount")))
	})
}

func TestCloneBackendIsStoredForMount(t *testing.T) {
	u := gptest.NewUnitTester(t)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	cfg := config.NewInMemory()
	require.NoError(t, cfg.SetPath(u.StoreDir("")))

	act, err := newAction(cfg, semver.Version{}, false)
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	c := gptest.CliCtx(ctx, t)
	_, err = act.IsInitialized(ctx, c)
	require.NoError(t, err)

	repo := aGitRepo(ctx, t, u, "my-project")

	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"check-keys": "false"}, repo, "the-project")
	require.NoError(t, act.Clone(ctx, c))

	require.Contains(t, act.cfg.Mounts(), "the-project")
}

func TestCloneGetGitConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)

	r1 := gptest.UnsetVars(termio.NameVars...)
	defer r1()
	r2 := gptest.UnsetVars(termio.EmailVars...)
	defer r2()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	name, email, err := act.cloneGetGitConfig(ctx, "foobar")
	require.NoError(t, err)
	assert.Equal(t, "0xDEADBEEF", name)
	assert.Equal(t, "0xDEADBEEF", email)
}

func TestCloneCheckDecryptionKeys(t *testing.T) {
	u := gptest.NewUnitTester(t)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	cfg := config.NewInMemory()
	require.NoError(t, cfg.SetPath(u.StoreDir("")))

	act, err := newAction(cfg, semver.Version{}, false)
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	c := gptest.CliCtx(ctx, t)
	_, err = act.IsInitialized(ctx, c)
	require.NoError(t, err)

	repo := aGitRepo(ctx, t, u, "my-project")

	if runtime.GOOS != "linux" {
		t.Skip("TODO: not working on non-linux builders, yet")
	}

	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"check-keys": "true"}, repo, "the-project")
	require.NoError(t, act.Clone(ctx, c))

	require.Contains(t, act.cfg.Mounts(), "the-project")
}

func TestHaveDecryptionKey(t *testing.T) {
	t.Parallel()

	// fpr models gopass's crypto.Fingerprint: it resolves any GPG-accepted form
	// of a known key (short ID, long ID, fingerprint or email) to the same
	// fingerprint, and returns an empty string for unknown keys.
	const fingerprint = "1234567890ABCDEF1234567890AB17F3ED51DADD9393"
	aliases := map[string]string{
		"0x17F3ED51DADD9393":                          fingerprint,                                   // short
		"0x7890AB17F3ED51DADD9393":                    fingerprint,                                   // long
		fingerprint:                                   fingerprint,                                   // full fingerprint
		"0x" + fingerprint:                            fingerprint,                                   // 0x-prefixed fingerprint
		"someone@example.org":                         fingerprint,                                   // email
		"1111111111111111111111111111111111111111111": "1111111111111111111111111111111111111111111", // unrelated key
	}
	fpr := func(_ context.Context, id string) string {
		return aliases[id]
	}

	// Our usable key, always reported in short form by ListIdentities.
	ids := []string{"0x17F3ED51DADD9393"}

	for _, tc := range []struct {
		name       string
		recipients []string
		want       bool
	}{
		{name: "short form recipient", recipients: []string{"0x17F3ED51DADD9393"}, want: true},
		{name: "long form recipient", recipients: []string{"0x7890AB17F3ED51DADD9393"}, want: true},
		{name: "fingerprint recipient", recipients: []string{fingerprint}, want: true},
		{name: "0x fingerprint recipient", recipients: []string{"0x" + fingerprint}, want: true},
		{name: "email recipient", recipients: []string{"someone@example.org"}, want: true},
		{name: "mixed recipients with a match", recipients: []string{"1111111111111111111111111111111111111111111", "someone@example.org"}, want: true},
		{name: "non-recipient key is rejected", recipients: []string{"1111111111111111111111111111111111111111111"}, want: false},
		{name: "no recipients", recipients: nil, want: false},
		{name: "unknown recipient", recipients: []string{"0xUNKNOWN"}, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := haveDecryptionKey(context.Background(), fpr, tc.recipients, ids)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHaveDecryptionKeyNoIdentities(t *testing.T) {
	t.Parallel()

	fpr := func(_ context.Context, id string) string { return "FP-" + id }

	assert.False(t, haveDecryptionKey(context.Background(), fpr, []string{"a", "b"}, nil))
}
