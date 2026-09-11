package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"

	"golang.org/x/crypto/hkdf"
)

func TestPlainRoundTrip(t *testing.T) {
	sess, out, err := New(AlgorithmPlain, nil)
	if err != nil {
		t.Fatalf("New(plain): %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("plain output = %q, want empty", out)
	}

	params, ciphertext, err := sess.Encrypt([]byte("hunter2"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(params) != 0 {
		t.Fatalf("plain params = %q, want empty", params)
	}
	if string(ciphertext) != "hunter2" {
		t.Fatalf("plain ciphertext = %q, want %q", ciphertext, "hunter2")
	}

	plaintext, err := sess.Decrypt(params, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(plaintext) != "hunter2" {
		t.Fatalf("plaintext = %q, want %q", plaintext, "hunter2")
	}
}

func TestUnsupportedAlgorithm(t *testing.T) {
	if _, _, err := New("rot13", nil); err == nil {
		t.Fatal("New(rot13) succeeded, want error")
	}
}

func TestDHRoundTrip(t *testing.T) {
	// Simulate the client side of the exchange.
	clientPriv, err := rand.Int(rand.Reader, dhPrime)
	if err != nil {
		t.Fatalf("client private key: %v", err)
	}
	clientPub := new(big.Int).Exp(dhGenerator, clientPriv, dhPrime)

	// Server side.
	sess, serverPub, err := New(AlgorithmDHAES, leftPad(clientPub.Bytes(), dhKeyLen))
	if err != nil {
		t.Fatalf("New(dh): %v", err)
	}

	// Client derives the same key from the server's public key.
	serverPubInt := new(big.Int).SetBytes(serverPub)
	shared := leftPad(new(big.Int).Exp(serverPubInt, clientPriv, dhPrime).Bytes(), dhKeyLen)
	reader := hkdf.New(sha256.New, shared, nil, nil)
	clientKey := make([]byte, 16)
	if _, err := reader.Read(clientKey); err != nil {
		t.Fatalf("client hkdf: %v", err)
	}

	// Server encrypts; client decrypts.
	params, ciphertext, err := sess.Encrypt([]byte("a very secret value"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(params) != aes.BlockSize {
		t.Fatalf("IV length = %d, want %d", len(params), aes.BlockSize)
	}

	block, err := aes.NewCipher(clientKey)
	if err != nil {
		t.Fatalf("client cipher: %v", err)
	}
	decrypted := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, params).CryptBlocks(decrypted, ciphertext)
	padLen := int(decrypted[len(decrypted)-1])
	got := decrypted[:len(decrypted)-padLen]
	if string(got) != "a very secret value" {
		t.Fatalf("client decrypted = %q, want %q", got, "a very secret value")
	}

	// Client encrypts; server decrypts.
	clientIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(clientIV); err != nil {
		t.Fatalf("client IV: %v", err)
	}
	clientPlain := []byte("reply from client")
	clientPad := aes.BlockSize - (len(clientPlain) % aes.BlockSize)
	clientPadded := append(append([]byte{}, clientPlain...), bytes.Repeat([]byte{byte(clientPad)}, clientPad)...)
	clientCT := make([]byte, len(clientPadded))
	cipher.NewCBCEncrypter(block, clientIV).CryptBlocks(clientCT, clientPadded)

	plaintext, err := sess.Decrypt(clientIV, clientCT)
	if err != nil {
		t.Fatalf("server Decrypt: %v", err)
	}
	if string(plaintext) != "reply from client" {
		t.Fatalf("server plaintext = %q, want %q", plaintext, "reply from client")
	}
}

func TestDHPadsShortSharedSecret(t *testing.T) {
	// A shared secret whose big.Int representation is shorter than 128 bytes
	// must be left-padded, not right-padded or truncated.
	short := []byte{0x01, 0x02, 0x03}
	padded := leftPad(short, dhKeyLen)
	if len(padded) != dhKeyLen {
		t.Fatalf("len(padded) = %d, want %d", len(padded), dhKeyLen)
	}
	if !bytes.Equal(padded[dhKeyLen-len(short):], short) {
		t.Fatalf("padded tail = %x, want %x", padded[dhKeyLen-len(short):], short)
	}
	for i := range dhKeyLen - len(short) {
		if padded[i] != 0 {
			t.Fatalf("padding byte %d = %d, want 0", i, padded[i])
		}
	}
}

func TestDHRejectsBadIV(t *testing.T) {
	clientPriv, _ := rand.Int(rand.Reader, dhPrime)
	clientPub := new(big.Int).Exp(dhGenerator, clientPriv, dhPrime)
	sess, _, err := New(AlgorithmDHAES, leftPad(clientPub.Bytes(), dhKeyLen))
	if err != nil {
		t.Fatalf("New(dh): %v", err)
	}
	if _, err := sess.Decrypt([]byte("short"), []byte("0123456789abcdef")); err == nil {
		t.Fatal("Decrypt accepted a short IV, want error")
	}
}
