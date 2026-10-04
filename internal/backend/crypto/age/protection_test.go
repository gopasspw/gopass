package age

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	ageapi "filippo.io/age"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func protectedKeyringContext(t *testing.T, identityFile, recipients string) context.Context {
	t.Helper()
	oldInput := termio.Stdin
	termio.Stdin = strings.NewReader(strings.Repeat("y\n", 20))
	t.Cleanup(func() { termio.Stdin = oldInput })
	cfg := config.NewInMemory()
	require.NoError(t, cfg.SetEnv("age.agent-enabled", "false"))
	require.NoError(t, cfg.SetEnv("age.keyring-identities", identityFile))
	require.NoError(t, cfg.SetEnv("age.keyring-recipients", recipients))

	return cfg.WithConfig(ctxutil.WithTerminal(ctxutil.WithInteractive(t.Context(), true), true))
}

func keyringTestIdentity(t *testing.T) *ageapi.X25519Identity {
	t.Helper()
	id, err := ageapi.GenerateX25519Identity()
	require.NoError(t, err)

	return id
}

func writeBootstrapIdentity(t *testing.T, path string, id *ageapi.X25519Identity) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(id.String()+"\n"), 0o600))
}

func readProtectedKeyring(t *testing.T, path string, id ageapi.Identity) string {
	t.Helper()
	ciphertext, err := os.ReadFile(path)
	require.NoError(t, err)
	r, err := ageapi.Decrypt(bytes.NewReader(ciphertext), id)
	require.NoError(t, err)
	plaintext, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(plaintext)
}

func TestRecipientProtectedKeyringRoundTrip(t *testing.T) {
	a := newTestAge(t)
	wrapper := keyringTestIdentity(t)
	software := keyringTestIdentity(t)
	bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
	writeBootstrapIdentity(t, bootstrap, wrapper)
	ctx := protectedKeyringContext(t, bootstrap, wrapper.Recipient().String())
	a.SetPasswordCallback(func(string, bool) ([]byte, error) {
		t.Error("recipient-protected keyring must not request a passphrase")

		return nil, assert.AnError
	})

	require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))
	assert.Equal(t, software.String(), readProtectedKeyring(t, a.identity, wrapper))
	ids, err := a.Identities(ctx)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, software.Recipient().String(), recipientOf(ids[0]))
	assert.NotEqual(t, wrapper.Recipient().String(), recipientOf(ids[0]), "bootstrap credential is not a store identity")
	ciphertext, err := os.ReadFile(a.identity)
	require.NoError(t, err)
	assert.NotContains(t, string(ciphertext), software.String())
	info, err := os.Stat(a.identity)
	require.NoError(t, err)
	// Windows FileMode exposes only the read-only attribute, not Unix mode bits.
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestRecipientProtectedKeyringMigrationPreservesRawIdentities(t *testing.T) {
	a := newTestAge(t)
	originalCtx := protectedKeyringContext(t, "", "")
	software := keyringTestIdentity(t)
	const pluginLine = "AGE-PLUGIN-YUBIKEY-1GKZKJQYZL98RLMC67F9PJ|age1yubikey1qt2r3tfk7wvlykudm7ew28dqqm3h8ln9zfsxsq4lcd2w8rh4n4hhz46ur24"
	raw := "# preserved comment\n" + software.String() + "\n" + pluginLine
	require.NoError(t, a.saveIdentities(originalCtx, strings.Split(raw, "\n"), true))
	wrapper := keyringTestIdentity(t)
	bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
	writeBootstrapIdentity(t, bootstrap, wrapper)
	protectedCtx := protectedKeyringContext(t, bootstrap, wrapper.Recipient().String())
	// Selecting a new write protection mode must not prevent unlocking the old scrypt format.
	got, err := a.loadIdentityFile(protectedCtx)
	require.NoError(t, err)
	assert.Equal(t, raw, got)
	require.NoError(t, a.reencryptIdentities(protectedCtx))
	assert.Equal(t, raw, readProtectedKeyring(t, a.identity, wrapper))

	added := keyringTestIdentity(t)
	require.NoError(t, a.addIdentity(protectedCtx, added))
	got = readProtectedKeyring(t, a.identity, wrapper)
	assert.Equal(t, raw+"\n"+added.String(), got, "mutation preserves raw plugin lines without invoking their plugins")
	// Rewriting the raw identities, as remove does, must retain recipient encryption.
	require.NoError(t, a.saveIdentities(protectedCtx, strings.Split(raw, "\n"), false))
	assert.Equal(t, raw, readProtectedKeyring(t, a.identity, wrapper))
}

func TestRecipientProtectedKeyringErrorsPreserveCiphertext(t *testing.T) {
	a := newTestAge(t)
	wrapper := keyringTestIdentity(t)
	software := keyringTestIdentity(t)
	bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
	writeBootstrapIdentity(t, bootstrap, wrapper)
	ctx := protectedKeyringContext(t, bootstrap, wrapper.Recipient().String())
	require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))
	original, err := os.ReadFile(a.identity)
	require.NoError(t, err)
	wrong := filepath.Join(t.TempDir(), "wrong.txt")
	writeBootstrapIdentity(t, wrong, keyringTestIdentity(t))
	for _, tc := range []struct{ name, path, recipients string }{
		{"missing bootstrap", filepath.Join(t.TempDir(), "missing.txt"), wrapper.Recipient().String()},
		{"wrong bootstrap", wrong, wrapper.Recipient().String()},
		{"circular bootstrap", a.identity, wrapper.Recipient().String()},
		{"invalid recipient", bootstrap, "not-an-age-recipient"},
		{"refuse silent downgrade", bootstrap, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			badCtx := protectedKeyringContext(t, tc.path, tc.recipients)
			if tc.name != "invalid recipient" && tc.name != "refuse silent downgrade" {
				_, err := a.Identities(badCtx)
				require.Error(t, err)
			}
			require.Error(t, a.reencryptIdentities(badCtx))
			current, err := os.ReadFile(a.identity)
			require.NoError(t, err)
			assert.Equal(t, original, current, "failed migration must leave the existing ciphertext untouched")
		})
	}
}

func TestRecipientProtectedKeyringRecoveryRecipient(t *testing.T) {
	a := newTestAge(t)
	hardware := keyringTestIdentity(t)
	recovery := keyringTestIdentity(t)
	software := keyringTestIdentity(t)
	bootstrap := filepath.Join(t.TempDir(), "unlock.txt")
	writeBootstrapIdentity(t, bootstrap, hardware)
	ctx := protectedKeyringContext(t, bootstrap, hardware.Recipient().String()+","+recovery.Recipient().String())
	require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))
	assert.Equal(t, software.String(), readProtectedKeyring(t, a.identity, hardware))
	assert.Equal(t, software.String(), readProtectedKeyring(t, a.identity, recovery))
	writeBootstrapIdentity(t, bootstrap, recovery)
	ids, err := a.Identities(ctx)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, software.Recipient().String(), recipientOf(ids[0]))
}
