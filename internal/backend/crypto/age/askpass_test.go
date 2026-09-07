package age

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestNewAskPass_KeychainDisabled(t *testing.T) {
	keyring.MockInit()

	ctx := config.NewContextInMemory()
	a := newAskPass(ctx)

	require.NotNil(t, a)
	_, isOsKeyring := a.cache.(*osKeyring)
	assert.False(t, isOsKeyring, "cache should be in-memory when keychain is disabled")
}

func TestNewAskPass_KeychainEnabled_KeyringAvailable(t *testing.T) {
	keyring.MockInit()

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "age.usekeychain", "true"))
	ctx := cfg.WithConfig(t.Context())

	a := newAskPass(ctx)

	require.NotNil(t, a)
	_, isOsKeyring := a.cache.(*osKeyring)
	assert.True(t, isOsKeyring, "cache should be *osKeyring when keychain is enabled and available")
}

func TestNewAskPass_KeychainEnabled_KeyringUnavailable(t *testing.T) {
	keyring.MockInitWithError(errors.New("no keyring"))
	keyringWarningOnce = sync.Once{}

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "age.usekeychain", "true"))
	ctx := cfg.WithConfig(t.Context())

	a := newAskPass(ctx)

	require.NotNil(t, a)
	_, isOsKeyring := a.cache.(*osKeyring)
	assert.False(t, isOsKeyring, "cache should fall back to in-memory when keyring is unavailable")
}

func TestNew_KeychainEnabled_KeyringUnavailable_WarnsUser(t *testing.T) {
	keyring.MockInitWithError(errors.New("no keyring"))
	keyringWarningOnce = sync.Once{}

	var buf bytes.Buffer
	oldStderr := out.Stderr
	out.Stderr = &buf
	t.Cleanup(func() { out.Stderr = oldStderr })

	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "age.usekeychain", "true"))
	ctx := cfg.WithConfig(t.Context())

	a, err := New(ctx, false, "")
	require.NoError(t, err)
	require.NotNil(t, a)

	assert.Contains(t, buf.String(), "OS keyring is not available")
}

func TestOsKeyring_Set_Success(t *testing.T) {
	keyring.MockInit()

	o := newOsKeyring()
	o.Set(t.Context(), "test-key", "test-value")

	val, err := keyring.Get("gopass", "test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", val)
}

func TestOsKeyring_Set_Failure(t *testing.T) {
	keyring.MockInitWithError(errors.New("no system keyring"))

	var buf bytes.Buffer
	oldStderr := out.Stderr
	out.Stderr = &buf
	t.Cleanup(func() { out.Stderr = oldStderr })

	o := newOsKeyring()
	o.Set(t.Context(), "test-key", "test-value")

	assert.Contains(t, buf.String(), "Failed to cache passphrase in OS keyring")
}

func TestOsKeyring_Get_Success(t *testing.T) {
	keyring.MockInit()

	// Pre-populate the keyring
	require.NoError(t, keyring.Set("gopass", "my-key", "my-secret"))

	o := newOsKeyring()
	val, found := o.Get("my-key")

	assert.True(t, found)
	assert.Equal(t, "my-secret", val)
}

func TestOsKeyring_Get_Failure(t *testing.T) {
	keyring.MockInit()
	// Don't pre-populate — key doesn't exist

	o := newOsKeyring()
	val, found := o.Get("nonexistent-key")

	assert.False(t, found)
	assert.Empty(t, val)
}

func TestNew_UsesContextPasswordCallback(t *testing.T) {
	keyring.MockInit()

	var called bool
	var purged string

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAgePassphrase(ctx, "from-env")
	ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, confirm bool) ([]byte, error) {
		called = true
		assert.Equal(t, "prompt", prompt)
		assert.True(t, confirm)

		return []byte("from-callback"), nil
	})
	ctx = ctxutil.WithPasswordPurgeCallback(ctx, func(prompt string) {
		purged = prompt
	})

	a, err := New(ctx, false, "")
	require.NoError(t, err)

	pw, err := a.effectivePwCallback(ctx, "ignored")("prompt", true)
	require.NoError(t, err)
	assert.Equal(t, []byte("from-callback"), pw)
	assert.True(t, called)

	a.effectivePwPurgeCallback()("prompt")
	assert.Equal(t, "prompt", purged)
}
