package age

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-pinentry/v4"
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

func TestNew_ContextPasswordCallbackOverridesAgePassphrase(t *testing.T) {
	keyring.MockInit()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAgePassphrase(ctx, "from-env")
	ctx = ctxutil.WithPasswordCallback(ctx, func(string, bool) ([]byte, error) {
		return []byte("from-callback"), nil
	})

	a, err := New(ctx, false, "")
	require.NoError(t, err)

	pw, err := a.effectivePwCallback(ctx, "ignored")("prompt", false)
	require.NoError(t, err)
	assert.Equal(t, []byte("from-callback"), pw)
}

type recordingPinentryProcess struct {
	written []string
}

func (r *recordingPinentryProcess) Start(string, []string) error {
	return nil
}

func (r *recordingPinentryProcess) Write(data []byte) (int, error) {
	r.written = append(r.written, strings.TrimRight(string(data), "\r\n"))

	return len(data), nil
}

func (r *recordingPinentryProcess) ReadLine() ([]byte, bool, error) {
	return []byte("OK"), false, nil
}

func (r *recordingPinentryProcess) Close() error {
	return nil
}

func TestPinentryTTYOptionsFor(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		validTTYs map[string]bool
		hasDevTTY bool
		wantCmds  []string
	}{
		{
			name:      "GPG_TTY unset falls back to /dev/tty",
			env:       map[string]string{"TERM": "xterm-256color", "LANG": "en_US.UTF-8"},
			hasDevTTY: true,
			wantCmds: []string{
				"OPTION ttyname=/dev/tty",
				"OPTION ttytype=xterm-256color",
				"OPTION lc-ctype=en_US.UTF-8",
			},
		},
		{
			name:      "valid GPG_TTY takes precedence over /dev/tty",
			env:       map[string]string{"GPG_TTY": "/dev/pts/4", "TERM": "xterm"},
			validTTYs: map[string]bool{"/dev/pts/4": true},
			hasDevTTY: true,
			wantCmds: []string{
				"OPTION ttyname=/dev/pts/4",
				"OPTION ttytype=xterm",
			},
		},
		{
			name:      "invalid GPG_TTY falls back to /dev/tty",
			env:       map[string]string{"GPG_TTY": "not a tty"},
			hasDevTTY: true,
			wantCmds: []string{
				"OPTION ttyname=/dev/tty",
			},
		},
		{
			name:      "headless without GPG_TTY or /dev/tty sends no ttyname",
			env:       map[string]string{"TERM": "dumb"},
			hasDevTTY: false,
			wantCmds: []string{
				"OPTION ttytype=dumb",
			},
		},
		{
			name:      "invalid GPG_TTY without /dev/tty sends no ttyname",
			env:       map[string]string{"GPG_TTY": "not a tty"},
			hasDevTTY: false,
			wantCmds:  nil,
		},
		{
			name: "LC_ALL takes precedence over LC_CTYPE and LANG",
			env: map[string]string{
				"LC_ALL":   "C.UTF-8",
				"LC_CTYPE": "de_DE.UTF-8",
				"LANG":     "en_US.UTF-8",
			},
			hasDevTTY: false,
			wantCmds: []string{
				"OPTION lc-ctype=C.UTF-8",
			},
		},
		{
			name: "LC_CTYPE takes precedence over LANG",
			env: map[string]string{
				"LC_CTYPE": "de_DE.UTF-8",
				"LANG":     "en_US.UTF-8",
			},
			hasDevTTY: false,
			wantCmds: []string{
				"OPTION lc-ctype=de_DE.UTF-8",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proc := &recordingPinentryProcess{}
			opts := append(
				[]pinentry.ClientOption{pinentry.WithProcess(proc)},
				pinentryTTYOptionsFor(
					func(k string) string { return tc.env[k] },
					func(p string) bool { return tc.validTTYs[p] },
					func() bool { return tc.hasDevTTY },
				)...,
			)

			client, err := pinentry.NewClient(opts...)
			require.NoError(t, err)
			require.NoError(t, client.Close())

			// Strip the trailing BYE sent by Close().
			require.NotEmpty(t, proc.written)
			require.Equal(t, "BYE", proc.written[len(proc.written)-1])
			got := proc.written[:len(proc.written)-1]
			if len(tc.wantCmds) == 0 {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tc.wantCmds, got)
			}
		})
	}
}

func TestIsCharDevice(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "regular")
	require.NoError(t, os.WriteFile(tmpFile, []byte("x"), 0o600))

	assert.False(t, isCharDevice(tmpFile))
	assert.False(t, isCharDevice(filepath.Join(t.TempDir(), "does-not-exist")))
	assert.False(t, isCharDevice("not a tty"))

	if runtime.GOOS != "windows" {
		assert.True(t, isCharDevice("/dev/null"))
	}
}
