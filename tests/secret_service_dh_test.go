//go:build linux

package tests

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/gopasspw/gopass/internal/secretservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/hkdf"
)

// The RFC 3526 MODP group 2 parameters, duplicated here so that the test acts
// as an independent DH client. The server-side implementation lives in
// internal/secretservice/crypto/dh.go.
var (
	testDHPrime = func() *big.Int {
		p, _ := new(big.Int).SetString(
			"FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD1"+
				"29024E088A67CC74020BBEA63B139B22514A08798E3404DD"+
				"EF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245"+
				"E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7ED"+
				"EE386BFB5A899FA5AE9F24117C4B1FE649286651ECE65381"+
				"FFFFFFFFFFFFFFFF", 16,
		)

		return p
	}()
	testDHGen = big.NewInt(2)
)

// testLeftPad left-pads b to n bytes.
func testLeftPad(b []byte, n int) []byte {
	if len(b) >= n {
		return b
	}

	out := make([]byte, n)
	copy(out[n-len(b):], b)

	return out
}

// testDHKey derives the AES-128 key the way a libsecret client would.
func testDHKey(serverPublic []byte, clientPrivate *big.Int) []byte {
	shared := new(big.Int).Exp(new(big.Int).SetBytes(serverPublic), clientPrivate, testDHPrime)
	reader := hkdf.New(sha256.New, testLeftPad(shared.Bytes(), 128), nil, nil)
	key := make([]byte, 16)

	if _, err := reader.Read(key); err != nil {
		panic(err)
	}

	return key
}

// testEncrypt encrypts plaintext with AES-128-CBC and PKCS#7, returning the IV
// and the ciphertext.
func testEncrypt(t *testing.T, key, plaintext []byte) ([]byte, []byte) {
	t.Helper()

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}

	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)

	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("rand: %v", err)
	}

	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	return iv, ciphertext
}

// testDecrypt decrypts an AES-128-CBC ciphertext and strips PKCS#7 padding.
func testDecrypt(t *testing.T, key, iv, ciphertext []byte) []byte {
	t.Helper()

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("aes.NewCipher: %v", err)
	}

	decrypted := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(decrypted, ciphertext)

	if len(decrypted) == 0 {
		t.Fatal("empty plaintext")
	}

	padLen := int(decrypted[len(decrypted)-1])
	if padLen == 0 || padLen > aes.BlockSize || padLen > len(decrypted) {
		t.Fatalf("invalid padding length: %d", padLen)
	}

	return decrypted[:len(decrypted)-padLen]
}

// TestSecretServiceDHRoundTrip exercises the
// "dh-ietf1024-sha256-aes128-cbc-pkcs7" transport algorithm end to end, acting
// as an independent libsecret-style client.
func TestSecretServiceDHRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	address, stopBus := startPrivateBus(t)
	defer stopBus()

	t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)

	logs := startSecretServiceDaemon(t, ts)

	conn := connectBus(t, address)
	waitForSecretService(t, conn, logs)

	svc := conn.Object(secretservice.ServiceName, secretservice.ServicePath)

	// Client side of the DH key exchange.
	clientPrivate, err := rand.Int(rand.Reader, testDHPrime)
	require.NoError(t, err)

	clientPublic := new(big.Int).Exp(testDHGen, clientPrivate, testDHPrime)

	var (
		output  dbus.Variant
		session dbus.ObjectPath
	)
	require.NoError(t, svc.Call(secretservice.ServiceIface+".OpenSession", 0, secretservice.AlgorithmDHAES, dbus.MakeVariant(testLeftPad(clientPublic.Bytes(), 128))).Store(&output, &session))
	require.NotEqual(t, secretservice.NullPath, session)

	t.Cleanup(func() {
		_ = conn.Object(secretservice.ServiceName, session).Call(secretservice.SessionIface+".Close", 0).Err
	})

	serverPublic, ok := output.Value().([]byte)
	require.True(t, ok, "server output is not a byte slice: %T", output.Value())
	require.Len(t, serverPublic, 128, "server public key must be left-padded to 128 bytes")

	key := testDHKey(serverPublic, clientPrivate)

	// Store an encrypted secret.
	var defaultPath dbus.ObjectPath
	require.NoError(t, svc.Call(secretservice.ServiceIface+".ReadAlias", 0, "default").Store(&defaultPath))

	coll := conn.Object(secretservice.ServiceName, defaultPath)

	iv, ciphertext := testEncrypt(t, key, []byte("dh-secret-value"))

	secret := secretservice.Secret{
		Session:     session,
		Parameters:  iv,
		Value:       ciphertext,
		ContentType: "text/plain",
	}
	props := map[string]dbus.Variant{
		secretservice.ItemIface + ".Label": dbus.MakeVariant("DH Item"),
	}

	var itemPath, promptPath dbus.ObjectPath
	require.NoError(t, coll.Call(secretservice.CollectionIface+".CreateItem", 0, props, secret, false).Store(&itemPath, &promptPath))
	require.NotEqual(t, secretservice.NullPath, itemPath)

	// Read it back and decrypt with the client key.
	var got secretservice.Secret
	require.NoError(t, conn.Object(secretservice.ServiceName, itemPath).Call(secretservice.ItemIface+".GetSecret", 0, session).Store(&got))
	assert.Len(t, got.Parameters, aes.BlockSize, "IV must be 16 bytes")
	assert.Equal(t, "dh-secret-value", string(testDecrypt(t, key, got.Parameters, got.Value)))

	// A GetSecrets batch call must use the same transport encryption.
	var batch map[dbus.ObjectPath]secretservice.Secret
	require.NoError(t, svc.Call(secretservice.ServiceIface+".GetSecrets", 0, []dbus.ObjectPath{itemPath}, session).Store(&batch))
	require.Contains(t, batch, itemPath)
	assert.Equal(t, "dh-secret-value", string(testDecrypt(t, key, batch[itemPath].Parameters, batch[itemPath].Value)))
}
