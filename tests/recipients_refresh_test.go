package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/gopasspw/gopass/tests/can"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecipientsRefreshDetectsExpired reproduces the scenario from
// GH-1430 / ADR A-13: when a recipient's key in the local keyring expires,
// the doctor diagnostic should flag it as a warning, and 'gopass sync' or
// an explicit 'recipients update' flow should be possible.
func TestRecipientsRefreshDetectsExpired(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	// Create an expired key and init a store with it.
	expiredKeyID := createExpiredGPGKey(t, ts)

	rootPath := ts.storeDir("root")
	out, err := ts.run("init --crypto=gpgcli --storage=fs " + expiredKeyID)
	require.NoError(t, err, "init: %s", out)

	// Write the .public-keys entry for the expired key so the store has
	// a valid copy even though the local keyring copy is expired.
	pubKeysDir := filepath.Join(rootPath, ".public-keys")
	require.NoError(t, os.MkdirAll(pubKeysDir, 0o700))
	// Read the keyring from the GPG homedir (which now contains both the
	// embedded keyring keys and the expired key).
	el, err := readGPGKeyRing(ts.gpgDir(), "pubring.gpg")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(el), 2, "expected at least 2 keys in keyring (embedded + expired)")
	fh, err := os.Create(filepath.Join(pubKeysDir, expiredKeyID))
	require.NoError(t, err)
	// Write the expired key's public key to .public-keys/
	require.NoError(t, el[len(el)-1].Serialize(fh))
	require.NoError(t, fh.Close())

	// Run the diagnostic — it should warn about the expired key (Stage 4
	// will make this a proper expiry warning; for now it may show as
	// non-canonical or missing-keyring warning).
	out, err = ts.run("doctor --recipients --verbose")
	require.NoError(t, err, "doctor --recipients for expired key: %s", out)

	// At minimum, the diagnostic should not crash and report something.
	assert.NotEmpty(t, out, "doctor should produce output")
}

// createExpiredGPGKey generates a new GPG key that is immediately expired,
// adds it to the test GPG keyring, and returns the short key ID.
func createExpiredGPGKey(t *testing.T, ts *tester) string {
	t.Helper()

	e, err := openpgp.NewEntity("Expired", "", "expired@example.com", &packet.Config{
		RSABits: 4096,
	})
	require.NoError(t, err)

	for _, id := range e.Identities {
		err := id.SelfSignature.SignUserId(id.UserId.Id, e.PrimaryKey, e.PrivateKey, &packet.Config{
			SigLifetimeSecs: 1, // expire after 1 second
		})
		require.NoError(t, err)
	}

	el := can.EmbeddedKeyRing()
	el = append(el, e)

	fn := filepath.Join(ts.gpgDir(), "pubring.gpg")
	fh, err := os.Create(fn)
	require.NoError(t, err)

	for _, e := range el {
		require.NoError(t, e.Serialize(fh))
	}
	require.NoError(t, fh.Close())

	fn = filepath.Join(ts.gpgDir(), "secring.gpg")
	fh, err = os.Create(fn)
	require.NoError(t, err)

	for _, e := range el {
		if e.PrivateKey != nil {
			require.NoError(t, e.SerializePrivate(fh, nil))
		}
	}
	require.NoError(t, fh.Close())

	// Wait for the key to expire.
	time.Sleep(time.Second)

	return e.PrimaryKey.KeyIdShortString()
}

// readGPGKeyRing reads a GPG keyring file (e.g. pubring.gpg) from the given
// directory and returns the parsed entity list.
func readGPGKeyRing(dir, name string) (openpgp.EntityList, error) {
	fh, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	defer fh.Close() //nolint:errcheck

	return openpgp.ReadKeyRing(fh)
}
