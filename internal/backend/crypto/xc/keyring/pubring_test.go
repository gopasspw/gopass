package keyring

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubring(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
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

	require.NoError(t, ioutil.WriteFile(fn, []byte("foobar"), 0644))
	kr, err := LoadPubring(fn, nil)
	assert.Error(t, err)
	assert.NoError(t, os.Remove(fn))
	assert.Nil(t, kr)

	kr, err = LoadPubring(fn, nil)
	assert.NoError(t, err)
	assert.NotNil(t, kr)

	kr.Set(&k1.PublicKey)
	kr.Set(&k2.PublicKey)

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

	buf, err := kr.Export(k2.Fingerprint())
	assert.NoError(t, err)

	assert.Equal(t, true, kr.Contains(k2.Fingerprint()))
	assert.NoError(t, kr.Remove(k2.Fingerprint()))
	assert.Error(t, kr.Remove(k2.Fingerprint()))
	assert.Equal(t, false, kr.Contains(k2.Fingerprint()))

	assert.NoError(t, kr.Import(buf))
	assert.Equal(t, true, kr.Contains(k2.Fingerprint()))
}
