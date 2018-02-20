package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var zeroArray32 = [32]uint8{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

func TestPrivateKeyDecrypt(t *testing.T) {
	passphrase := "test"

	key, err := GenerateKeypair(passphrase)
	assert.NoError(t, err)
	t.Logf("Key: %+v\n", key)

	assert.NoError(t, key.Encrypt(passphrase))
	t.Logf("Key: %+v\n", key)

	assert.NoError(t, key.Decrypt(passphrase))
	t.Logf("Key: %+v\n", key)
	assert.NotEqual(t, zeroArray32, key.privateKey)
}

func TestPrivateKeyMarshal(t *testing.T) {
	passphrase := "test"

	key, err := GenerateKeypair(passphrase)
	assert.NoError(t, err)

	assert.NoError(t, key.Encrypt(passphrase))
	t.Logf("Key: %+v\n", key)

	assert.NoError(t, key.Decrypt(passphrase))
	t.Logf("Key: %+v\n", key)
	assert.NotEqual(t, zeroArray32, key.PrivateKey())
}
