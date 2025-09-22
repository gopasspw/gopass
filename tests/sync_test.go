package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/gopasspw/gopass/tests/can"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("sync")
	require.NoError(t, err)
	assert.Contains(t, out, "All done")
}

func createGPGKey(t *testing.T, ts *tester, name, email string) string {
	t.Helper()
	e, err := openpgp.NewEntity(name, "", email, &packet.Config{
		RSABits: 4096,
	})
	require.NoError(t, err)

	for _, id := range e.Identities {
		err := id.SelfSignature.SignUserId(id.UserId.Id, e.PrimaryKey, e.PrivateKey, &packet.Config{})
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

	return e.PrimaryKey.KeyIdShortString()
}

func TestSyncKeepSubkey(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	// init store
	ts.initStore()

	// create a new key
	keyID := createGPGKey(t, ts, "sub-store-key", "sub-store-key@example.com")

	// create a secret in a subdirectory
	secretPath := "project_123/secret1"
	out, err := ts.run("insert -f " + secretPath)
	require.NoError(t, err, "failed to insert secret: %s", out)

	// create .gpg-id file
	subDir := filepath.Join(ts.storeDir("root"), "project_123")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	gpgIDFile := filepath.Join(subDir, ".gpg-id")
	err = os.WriteFile(gpgIDFile, []byte(keyID), 0o644)
	require.NoError(t, err)

	// re-encrypt the secret
	out, err = ts.run("fsck --decrypt " + filepath.Dir(secretPath))
	require.NoError(t, err, "failed to fsck: %s", out)

	// sync the store
	out, err = ts.run("sync")
	require.NoError(t, err, "failed to sync: %s", out)

	// export the public key
	pubKeyFile := filepath.Join(ts.storeDir("root"), ".public-keys", keyID)
	out, err = ts.runCmd([]string{"gpg", "--armor", "--export", keyID}, nil)
	require.NoError(t, err, "failed to export key: %s", out)
	err = os.WriteFile(pubKeyFile, []byte(out), 0o644)
	require.NoError(t, err)

	// run sync again
	out, err = ts.run("sync")
	require.NoError(t, err, "failed to sync: %s", out)

	// check if the public key file still exists
	assert.FileExists(t, pubKeyFile, "public key file should exist")
}
