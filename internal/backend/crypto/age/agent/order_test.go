package agent

import (
	"bytes"
	"errors"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/require"
)

// testPluginIdentity models a plugin identity without external binaries or
// authentication: it has a distinct concrete type and records Unwrap calls.
type testPluginIdentity struct {
	identity age.Identity
	calls    int
	err      error
}

func (i *testPluginIdentity) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	i.calls++
	if i.err != nil {
		return nil, i.err
	}

	return i.identity.Unwrap(stanzas)
}

func TestAgentDecryptIdentityOrder(t *testing.T) {
	native, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	pluginKey, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	// Both identities can decrypt this real ciphertext. Checking plaintext
	// alone would miss age moving the native identity ahead of the plugin.
	plaintext := []byte("identity order regression")
	var ciphertext bytes.Buffer
	writer, err := age.Encrypt(&ciphertext, native.Recipient(), pluginKey.Recipient())
	require.NoError(t, err)
	_, err = writer.Write(plaintext)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	t.Run("plugin first", func(t *testing.T) {
		plugin := &testPluginIdentity{identity: pluginKey}
		a := &Agent{identities: []age.Identity{plugin, native}}

		decrypted, err := a.decrypt(ciphertext.Bytes())
		require.NoError(t, err)
		require.Equal(t, plaintext, decrypted)
		require.Equal(t, 1, plugin.calls)
	})

	t.Run("native first", func(t *testing.T) {
		plugin := &testPluginIdentity{identity: pluginKey}
		a := &Agent{identities: []age.Identity{native, plugin}}

		decrypted, err := a.decrypt(ciphertext.Bytes())
		require.NoError(t, err)
		require.Equal(t, plaintext, decrypted)
		require.Zero(t, plugin.calls)
	})

	t.Run("nonmatching plugin falls back", func(t *testing.T) {
		nonmatching, err := age.GenerateX25519Identity()
		require.NoError(t, err)
		plugin := &testPluginIdentity{identity: nonmatching}
		a := &Agent{identities: []age.Identity{plugin, native}}

		decrypted, err := a.decrypt(ciphertext.Bytes())
		require.NoError(t, err)
		require.Equal(t, plaintext, decrypted)
		require.Equal(t, 1, plugin.calls)
	})

	t.Run("plugin error prevents fallback", func(t *testing.T) {
		authErr := errors.New("authentication cancelled")
		plugin := &testPluginIdentity{err: authErr}
		a := &Agent{identities: []age.Identity{plugin, native}}

		decrypted, err := a.decrypt(ciphertext.Bytes())
		require.ErrorIs(t, err, authErr)
		require.Nil(t, decrypted)
		require.Equal(t, 1, plugin.calls)
	})
}
