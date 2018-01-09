package keyring

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubring(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	fn := filepath.Join(td, "pubring.xcb")

	passphrase := "test"

	k1, err := GenerateKeypair(passphrase)
	assert.NoError(t, err)
	k2, err := GenerateKeypair(passphrase)
	assert.NoError(t, err)

	k1Fp := k1.PublicKey.Fingerprint()
	var k1Pk [32]byte
	copy(k1Pk[:], k1.PublicKey.PublicKey[:])

	kr, err := LoadPubring(fn, nil)
	assert.NoError(t, err)
	assert.NotNil(t, kr)

	assert.NoError(t, kr.Set(&k1.PublicKey))
	assert.NoError(t, kr.Set(&k2.PublicKey))

	assert.NoError(t, kr.Save())

	kr, err = LoadPubring(fn, nil)
	assert.NoError(t, err)
	assert.NotNil(t, kr)

	for _, key := range kr.KeyIDs() {
		pk := kr.Get(key)
		t.Logf("PublicKey: %+v", pk)
		if pk.Fingerprint() == k1Fp {
			assert.Equal(t, k1Fp, pk.PublicKey)
		}
	}

	assert.Equal(t, true, kr.Contains(k1.Fingerprint()))
	assert.Equal(t, true, kr.Contains(k2.Fingerprint()))
	assert.NoError(t, kr.Remove(k2.Fingerprint()))
	assert.Error(t, kr.Remove(k2.Fingerprint()))
}
