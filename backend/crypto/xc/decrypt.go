package xc

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/justwatchcom/gopass/backend/crypto/xc/keyring"
	"github.com/justwatchcom/gopass/backend/crypto/xc/xcpb"
	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/box"
)

const (
	maxUnlockAttempts = 3
)

// Decrypt tries to decrypt the given ciphertext and returns the plaintext
func (x *XC) Decrypt(ctx context.Context, buf []byte) ([]byte, error) {
	// unmarshal the protobuf message, the header and body are still encrypted
	// afterwards (parts of the header are plaintext!)
	msg := &xcpb.Message{}
	if err := proto.Unmarshal(buf, msg); err != nil {
		return nil, err
	}

	// try to find a suiteable decryption key in the header
	sk, err := x.decryptSessionKey(msg.Header)
	if err != nil {
		return nil, err
	}

	// initialize the AEAD cipher with the session key
	cp, err := chacha20poly1305.New(sk)
	if err != nil {
		return nil, err
	}

	// decrypt and verify the ciphertext
	plaintext, err := cp.Open(nil, msg.Header.Nonce, msg.Body, nil)
	if err != nil {
		return nil, err
	}

	if !msg.Compressed {
		return plaintext, nil
	}

	return decompress(plaintext)
}

// findDecryptionKey tries to find a suiteable decryption key from the available
// decryption keys and the recipients
func (x *XC) findDecryptionKey(hdr *xcpb.Header) (*keyring.PrivateKey, error) {
	for _, pk := range x.secring.KeyIDs() {
		if _, found := hdr.Recipients[pk]; found {
			return x.secring.Get(pk), nil
		}
	}
	return nil, fmt.Errorf("no decryption key found for: %+v", hdr.Recipients)
}

// findPublicKey tries to find a given public key in the keyring
func (x *XC) findPublicKey(needle string) (*keyring.PublicKey, error) {
	for _, id := range x.pubring.KeyIDs() {
		if id == needle {
			return x.pubring.Get(id), nil
		}
	}
	return nil, fmt.Errorf("no sender found for id '%s'", needle)
}

// decryptPrivateKey will ask the agent to unlock the private key
func (x *XC) decryptPrivateKey(recp *keyring.PrivateKey) error {
	fp := recp.Fingerprint()

	for i := 0; i < maxUnlockAttempts; i++ {
		// retry asking for key in case it's wrong
		passphrase, err := x.client.Passphrase(fp, fmt.Sprintf("Unlock private key %s", recp.Fingerprint()))
		if err != nil {
			return errors.Wrapf(err, "failed to get passphrase from agent: %s", err)
		}

		if err = recp.Decrypt(passphrase); err == nil {
			// passphrase is correct, the key should now be decrypted
			return nil
		}

		// decryption failed, clear cache and wait a moment before trying again
		if err := x.client.Remove(fp); err != nil {
			return errors.Wrapf(err, "failed to clear cache")
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("failed to unlock private key '%s' after %d retries", fp, maxUnlockAttempts)
}

// decryptSessionKey will attempt to find a readable recipient entry in the
// header and decrypt it's session key
func (x *XC) decryptSessionKey(hdr *xcpb.Header) ([]byte, error) {
	// find a suiteable decryption key, i.e. a recipient entry which was encrypted
	// for one of our private keys
	recp, err := x.findDecryptionKey(hdr)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find decryption key")
	}

	// we need the senders public key to decrypt/verify the message, since the
	// box algorithm ties successful decryption to successful verification
	sender, err := x.findPublicKey(hdr.Sender)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find sender pub key for signature verification: %s", hdr.Sender)
	}

	// unlock recipient key
	if err := x.decryptPrivateKey(recp); err != nil {
		return nil, err
	}

	// this is the per recipient ciphertext, we need to decrypt it to extract
	// the session key
	ciphertext := hdr.Recipients[recp.Fingerprint()]

	// since box works with byte arrays (or: pointers thereof) we need to copy
	// the slice to fixed arrays
	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])

	var privKey [32]byte
	pk := recp.PrivateKey()
	copy(privKey[:], pk[:])

	// now we can try to decrypt/verify the ciphertext. unfortunately box doesn't give
	// us any diagnostic information in case it fails, i.e. we can't discern between
	// a failed decryption and a failed verification
	decrypted, ok := box.Open(nil, ciphertext[24:], &nonce, &sender.PublicKey, &privKey)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt session key")
	}
	return decrypted, nil
}
