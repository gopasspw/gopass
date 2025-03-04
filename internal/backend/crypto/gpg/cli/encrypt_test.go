package cli

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg/gpgconf"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/tests/can"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	if testing.Short() || runtime.GOOS != "linux" { // not working on darwin right now, can't test on windows
		t.Skip("skipping test in short mode.")
	}

	// necessary for setting up the env
	u := gptest.NewGUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)

	g, err := New(ctx, Config{
		Umask: fsutil.Umask(),
		Args:  gpgconf.GPGOpts(),
	})
	require.NoError(t, err)

	// import keys so GPG4Win can find them
	el := can.EmbeddedKeyRing()
	for _, e := range el {
		buf := &bytes.Buffer{}
		require.NoError(t, e.Serialize(buf))

		require.NoError(t, g.ImportPublicKey(ctx, buf.Bytes()))
	}

	plaintext := []byte("plaintext")
	ciphertext, err := g.Encrypt(ctx, plaintext, []string{can.KeyID()})
	require.NoError(t, err)

	plaintext2, err := g.Decrypt(ctx, ciphertext)
	require.NoError(t, err)

	assert.Equal(t, plaintext, plaintext2)
}
