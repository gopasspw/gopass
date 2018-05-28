package keyring

import (
	crypto_rand "crypto/rand"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/xcpb"
)

// saltLength is chosen based on the recommendation in
// https://tools.ietf.org/html/draft-irtf-cfrg-argon2-03#section-3.1
const (
	saltLength  = 16
	nonceLength = 24
	keyLength   = 32
)

// PrivateKey is a private key part of a keypair
type PrivateKey struct {
	PublicKey
	Encrypted     bool
	EncryptedData []byte
	privateKey    [keyLength]byte   // only available after decryption
	Nonce         [nonceLength]byte // for private key encryption

	Salt []byte // for KDF
}

// PrivateKey returns the decrypted private key material
func (p *PrivateKey) PrivateKey() [keyLength]byte {
	return p.privateKey
}

// GenerateKeypair generates a new keypair
func GenerateKeypair(passphrase string) (*PrivateKey, error) {
	pub, priv, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return nil, err
	}
	k := &PrivateKey{
		PublicKey: PublicKey{
			CreationTime: time.Now(),
			PubKeyAlgo:   PubKeyNaCl,
			PublicKey:    *pub,
			Identity:     &xcpb.Identity{},
		},
		Encrypted:  true,
		privateKey: *priv,
	}
	err = k.Encrypt(passphrase)
	return k, err
}

// Encrypt encrypts the private key material with the given passphrase
func (p *PrivateKey) Encrypt(passphrase string) error {
	p.Salt = make([]byte, saltLength)
	if n, err := crypto_rand.Read(p.Salt); err != nil || n < len(p.Salt) {
		return err
	}
	secretKey := p.deriveKey(passphrase)

	var nonce [nonceLength]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return err
	}
	p.Nonce = nonce

	p.EncryptedData = secretbox.Seal(nil, p.privateKey[:], &nonce, &secretKey)
	return nil
}

// Decrypt decrypts the private key
func (p *PrivateKey) Decrypt(passphrase string) error {
	if !p.Encrypted {
		return nil
	}
	secretKey := p.deriveKey(passphrase)

	decrypted, ok := secretbox.Open(nil, p.EncryptedData, &p.Nonce, &secretKey)
	if !ok {
		return fmt.Errorf("decryption error")
	}
	copy(p.privateKey[:], decrypted)

	p.Encrypted = false
	return nil
}

func (p *PrivateKey) deriveKey(passphrase string) [keyLength]byte {
	secretKeyBytes := argon2.IDKey([]byte(passphrase), p.Salt, 4, 64*1024, 4, 32)
	var secretKey [keyLength]byte
	copy(secretKey[:], secretKeyBytes)
	return secretKey
}
