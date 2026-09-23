package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"golang.org/x/crypto/hkdf"
)

// dhPrime is the RFC 2409 / RFC 3526 MODP group 2 (1024-bit) prime. The two
// specifications define the same group for the
// "dh-ietf1024-sha256-aes128-cbc-pkcs7" algorithm.
var dhPrime = func() *big.Int {
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

// dhGenerator is the generator of the MODP group.
var dhGenerator = big.NewInt(2)

// dhKeyLen is the length of the padded DH values in bytes (1024 bits).
const dhKeyLen = 128

// dhSession is an established DH/AES session.
type dhSession struct {
	aesKey []byte
}

// newDHSession performs the server side of the DH key exchange and returns the
// session plus the server's public key.
func newDHSession(clientPublic []byte) (*dhSession, []byte, error) {
	// Generate the server's ephemeral private key.
	privateKey, err := rand.Int(rand.Reader, dhPrime)
	if err != nil {
		return nil, nil, fmt.Errorf("generate private key: %w", err)
	}

	// serverPub = g^private mod p
	publicKey := new(big.Int).Exp(dhGenerator, privateKey, dhPrime)

	// shared = clientPub^private mod p
	clientPub := new(big.Int).SetBytes(clientPublic)
	sharedSecret := new(big.Int).Exp(clientPub, privateKey, dhPrime)

	// The shared secret must be left-padded to exactly 128 bytes before key
	// derivation. big.Int.Bytes() strips leading zero bytes, which would
	// otherwise change the derived key for some exchanges.
	sharedBytes := leftPad(sharedSecret.Bytes(), dhKeyLen)

	// aesKey = HKDF-SHA256(shared, salt=nil, info=nil)[0:16].
	hkdfReader := hkdf.New(sha256.New, sharedBytes, nil, nil)
	aesKey := make([]byte, 16)
	if _, err := hkdfReader.Read(aesKey); err != nil {
		return nil, nil, fmt.Errorf("hkdf: %w", err)
	}

	return &dhSession{aesKey: aesKey}, leftPad(publicKey.Bytes(), dhKeyLen), nil
}

// leftPad returns b left-padded with zero bytes to exactly n bytes. If b is
// already n or more bytes it is returned unchanged.
func leftPad(b []byte, n int) []byte {
	if len(b) >= n {
		return b
	}

	out := make([]byte, n)
	copy(out[n-len(b):], b)

	return out
}

// Encrypt encrypts plaintext with AES-128-CBC and PKCS#7 padding. The returned
// params are a fresh random IV.
func (s *dhSession) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return nil, nil, err
	}

	// PKCS#7 padding.
	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, err
	}

	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	return iv, ciphertext, nil
}

// Decrypt decrypts an AES-128-CBC ciphertext whose IV is params.
func (s *dhSession) Decrypt(params, ciphertext []byte) ([]byte, error) {
	if len(params) != aes.BlockSize {
		return nil, fmt.Errorf("invalid IV length: %d", len(params))
	}
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid ciphertext length: %d", len(ciphertext))
	}

	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, params).CryptBlocks(decrypted, ciphertext)

	if len(decrypted) == 0 {
		return nil, fmt.Errorf("empty plaintext")
	}

	padLen := int(decrypted[len(decrypted)-1])
	if padLen == 0 || padLen > aes.BlockSize || padLen > len(decrypted) {
		return nil, fmt.Errorf("invalid padding length: %d", padLen)
	}
	for i := len(decrypted) - padLen; i < len(decrypted); i++ {
		if decrypted[i] != byte(padLen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return decrypted[:len(decrypted)-padLen], nil
}

// Close wipes the AES key.
func (s *dhSession) Close() error {
	for i := range s.aesKey {
		s.aesKey[i] = 0
	}

	return nil
}
