package age

import (
	"errors"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// interactiveIdentity models a plugin identity without requiring a hardware
// token or external plugin executable. It can unwrap a real encrypted file key.
type interactiveIdentity struct {
	age.Identity
	calls int
	err   error
}

func (id *interactiveIdentity) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	id.calls++
	if id.err != nil {
		return nil, id.err
	}

	return id.Identity.Unwrap(stanzas)
}

func TestDecryptIdentityOrder(t *testing.T) {
	t.Parallel()

	native, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	pluginKey, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	backend := &Age{}
	plaintext := []byte("identity order regression")
	// Both identities match the ciphertext. Native-first sorting would bypass
	// the preferred interactive identity even though it can decrypt this entry.
	ciphertext, err := backend.encrypt(plaintext, native.Recipient(), pluginKey.Recipient())
	require.NoError(t, err)
	fatal := errors.New("authentication cancelled")

	for _, tc := range []struct {
		name        string
		nativeFirst bool
		pluginError error
		wantCalls   int
		wantError   error
	}{
		{name: "interactive first", wantCalls: 1},
		{name: "native first", nativeFirst: true},
		{name: "nonmatching interactive falls back", pluginError: age.ErrIncorrectIdentity, wantCalls: 1},
		{name: "authentication error stops decryption", pluginError: fatal, wantCalls: 1, wantError: fatal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			interactive := &interactiveIdentity{Identity: pluginKey, err: tc.pluginError}
			ids := []age.Identity{interactive, native}
			if tc.nativeFirst {
				ids = []age.Identity{native, interactive}
			}

			got, err := backend.decrypt(ciphertext, ids...)
			if tc.wantError != nil {
				require.ErrorIs(t, err, tc.wantError)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, plaintext, got)
			}
			assert.Equal(t, tc.wantCalls, interactive.calls)
		})
	}
}

func TestDecryptScryptIdentity(t *testing.T) {
	t.Parallel()

	const passphrase = "test keyring passphrase"
	recipient, err := age.NewScryptRecipient(passphrase)
	require.NoError(t, err)
	// Keep the test inexpensive without changing production keyring settings.
	recipient.SetWorkFactor(10)
	id, err := age.NewScryptIdentity(passphrase)
	require.NoError(t, err)
	backend := &Age{}
	plaintext := []byte("test keyring identities")
	ciphertext, err := backend.encrypt(plaintext, recipient)
	require.NoError(t, err)
	got, err := backend.decrypt(ciphertext, id)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}
