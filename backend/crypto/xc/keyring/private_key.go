package keyring

import (
	"fmt"
	"io"
	"time"

	"github.com/justwatchcom/gopass/backend/crypto/xc/xcpb"

	crypto_rand "crypto/rand"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

// PrivateKey is a private key part of a keypair
type PrivateKey struct {
	PublicKey
	Encrypted     bool
	EncryptedData []byte
	privateKey    [32]byte // only available after decryption
	Nonce         [24]byte // for private key encryption

	Salt []byte // for KDF
}

// PrivateKey returns the decrypted private key material
func (p *PrivateKey) PrivateKey() [32]byte {
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
	p.Salt = make([]byte, 12)
	if n, err := crypto_rand.Read(p.Salt); err != nil || n < len(p.Salt) {
		return err
	}
	secretKey := p.deriveKey(passphrase)
	//fmt.Printf("[Encrypt] Passphrase: %s -> SecretKey: %x\n", passphrase, secretKey)
	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return err
	}
	//fmt.Printf("[Encrypt] Plaintext: %x\n", p.privateKey)
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
	//fmt.Printf("[Decrypt] Passphrase: %s -> SecretKey: %x\n", passphrase, secretKey)
	decrypted, ok := secretbox.Open(nil, p.EncryptedData, &p.Nonce, &secretKey)
	if !ok {
		return fmt.Errorf("decryption error")
	}
	copy(p.privateKey[:], decrypted)
	//fmt.Printf("[Decrypt] Plaintext: %x\n", p.privateKey)
	p.Encrypted = false
	return nil
}

func (p *PrivateKey) deriveKey(passphrase string) [32]byte {
	secretKeyBytes := argon2.Key([]byte(passphrase), p.Salt, 4, 32*1024, 4, 32)
	var secretKey [32]byte
	copy(secretKey[:], secretKeyBytes)
	return secretKey
}
