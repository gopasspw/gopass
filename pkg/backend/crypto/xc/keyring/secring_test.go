package keyring

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecring(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	fn := filepath.Join(td, "secring.xcb")

	passphrase := "test"

	k1, err := GenerateKeypair(passphrase)
	assert.NoError(t, err)
	k2, err := GenerateKeypair(passphrase)
	assert.NoError(t, err)

	kr, err := LoadSecring(fn)
	assert.NoError(t, err)
	assert.NotNil(t, kr)

	assert.NoError(t, kr.Set(k1))
	assert.NoError(t, kr.Set(k2))

	assert.NoError(t, kr.Save())

	kr, err = LoadSecring(fn)
	assert.NoError(t, err)
	assert.NotNil(t, kr)

	for _, key := range kr.KeyIDs() {
		pk := kr.Get(key)
		t.Logf("PrivateKey: %+v", pk)
		assert.Equal(t, true, pk.Encrypted)
		assert.NoError(t, pk.Decrypt(passphrase))
		assert.Equal(t, false, pk.Encrypted)
		t.Logf("PrivateKey: %+v", pk)
	}

	assert.Equal(t, true, kr.Contains(k1.Fingerprint()))

	buf, err := kr.Export(k2.Fingerprint(), true)
	assert.NoError(t, err)

	assert.Equal(t, true, kr.Contains(k2.Fingerprint()))
	assert.NoError(t, kr.Remove(k2.Fingerprint()))
	assert.Error(t, kr.Remove(k2.Fingerprint()))
	assert.Equal(t, false, kr.Contains(k2.Fingerprint()))

	assert.NoError(t, kr.Import(buf))
	assert.Equal(t, true, kr.Contains(k2.Fingerprint()))
}
