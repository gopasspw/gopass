package action

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupAgeGitFS(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Age)
	ctx = backend.WithStorageBackend(ctx, backend.GitFS)
	ctx = ctxutil.WithPasswordCallback(ctx, func(_ string, _ bool) ([]byte, error) {
		return []byte("foobar"), nil
	})
	ctx = ctxutil.WithPasswordPurgeCallback(ctx, func(s string) {})

	t.Skip("TODO: fix setup test")

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	// remove existing config and store, we want to start fresh
	assert.NoError(t, os.RemoveAll(u.StoreDir("")))
	assert.NoError(t, os.Remove(u.GPConfig()))

	c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"storage": "gitfs", "crypto": "age"})
	assert.Error(t, act.IsInitialized(c))
	assert.NoError(t, act.Setup(c))
	assert.Contains(t, buf.String(), "Welcome to gopass")

	crypto := act.Store.Crypto(ctx, "")
	require.NotNil(t, crypto)
	assert.Equal(t, "age", crypto.Name())
	assert.True(t, act.initHasUseablePrivateKeys(ctx, crypto))
	assert.Error(t, act.initGenerateIdentity(ctx, crypto, "foo bar", "foo.bar@example.org"))
	buf.Reset()

	act.printRecipients(ctx, "")
	assert.Contains(t, buf.String(), "0xDEADBEEF")
	buf.Reset()
}

func TestSetupPlainFS(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	c := gptest.CliCtx(ctx, t, "foo.bar@example.org")
	assert.NoError(t, act.IsInitialized(c))
	buf.Reset()

	t.Skip("TODO: fix these tests")

	assert.Error(t, act.Init(c))
	assert.Contains(t, buf.String(), "already initialized")
	buf.Reset()

	// this will abort because the store is already initialized
	assert.NoError(t, act.Setup(c))
	assert.Contains(t, buf.String(), "already initialized")
	buf.Reset()

	crypto := act.Store.Crypto(ctx, "")
	require.NotNil(t, crypto)
	assert.Equal(t, "plain", crypto.Name())
	assert.True(t, act.initHasUseablePrivateKeys(ctx, crypto))
	assert.Error(t, act.initGenerateIdentity(ctx, crypto, "foo bar", "foo.bar@example.org"))
	buf.Reset()

	act.printRecipients(ctx, "")
	assert.Contains(t, buf.String(), "0xDEADBEEF")
	buf.Reset()

	// un-initialize the store
	assert.NoError(t, os.Remove(filepath.Join(u.StoreDir(""), plain.IDFile)))
	assert.Error(t, act.IsInitialized(c))
	buf.Reset()

	// remove existing config and store
	assert.NoError(t, os.RemoveAll(u.StoreDir("")))
	assert.NoError(t, os.Remove(u.GPConfig()))

	// re-initialize the store, i.e. test that a fresh setup with plain and fs works
	assert.NoError(t, act.Setup(c))
	assert.Contains(t, buf.String(), "Welcome to gopass")
	buf.Reset()
}
