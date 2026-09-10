package root

import (
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)

	cfg := config.NewInMemory()
	require.NoError(t, cfg.SetPath(u.StoreDir("rs")))
	rs := New(cfg)

	inited, err := rs.IsInitialized(ctx)
	require.NoError(t, err)
	assert.False(t, inited)
	require.NoError(t, rs.Init(ctx, "", u.StoreDir("rs"), "0xDEADBEEF"))

	inited, err = rs.IsInitialized(ctx)
	require.NoError(t, err)
	assert.True(t, inited)
	require.NoError(t, rs.Init(ctx, "rs2", u.StoreDir("rs2"), "0xDEADBEEF"))
}

// TestInitializeDoesNotRewriteMountPaths guards against a regression where
// initializing the root store persisted the (normalized) mount paths back to
// the config on every invocation. Since gopass 1.17.2 mount paths are stored
// relative to the home directory (~/...), so re-writing an unshrunk absolute
// path from an existing config made gopass fail on read-only configs and
// spam errors like "failed to set mount path: failed to write config".
func TestInitializeDoesNotRewriteMountPaths(t *testing.T) {
	u := gptest.NewUnitTester(t)
	require.NoError(t, u.InitStore("sub1"))

	// mimic a pre-existing config that stores an absolute, not yet shrunk
	// mount path (as written by gopass releases before #3439)
	absPath := u.StoreDir("sub1")

	cfg := config.New()
	require.NoError(t, cfg.Set("", "mounts.sub1.path", absPath))

	// reload so the mount is picked up from the on-disk config
	cfg = config.New()
	require.Equal(t, absPath, cfg.Get("mounts.sub1.path"))

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)

	rs := New(cfg)

	// this triggers initialize(), which iterates over all configured mounts
	_, err := rs.IsInitialized(ctx)
	require.NoError(t, err)

	// the mount path must be left untouched
	assert.Equal(t, absPath, cfg.Get("mounts.sub1.path"))
}
