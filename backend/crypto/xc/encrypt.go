package xc

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/justwatchcom/gopass/backend/crypto/xc/keyring"
	"github.com/justwatchcom/gopass/backend/crypto/xc/xcpb"
	"github.com/pkg/errors"

	crypto_rand "crypto/rand"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/box"
)

const (
	// OnDiskVersion is the version of our on-disk format
	OnDiskVersion = 1
	chunkSizeMax  = 1024 * 1024
)

// Encrypt encrypts the given plaintext for all the given recipients and returns the
// ciphertext
func (x *XC) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	privKeyIDs := x.secring.KeyIDs()
	if len(privKeyIDs) < 1 {
		return nil, fmt.Errorf("no signing keys available on our keyring")
	}
	privKey := x.secring.Get(privKeyIDs[0])

	var compressed bool
	plaintext, compressed = compress(plaintext)

	// encrypt body (also generates a random session key)
	sk, chunks, err := encryptBody(plaintext)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encrypt body: %s", err)
	}

	// encrypt the session key per recipient
	header, err := x.encryptHeader(privKey, sk, recipients)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encrypt header: %s", err)
	}

	msg := &xcpb.Message{
		Version:    OnDiskVersion,
		Header:     header,
		Chunks:     chunks,
		Compressed: compressed,
	}

	return proto.Marshal(msg)
}

// encrypt header creates and populates a header struct with the nonce (plain)
// and the session key encrypted per recipient
func (x *XC) encryptHeader(signKey *keyring.PrivateKey, sk []byte, recipients []string) (*xcpb.Header, error) {
	hdr := &xcpb.Header{
		Sender:     signKey.Fingerprint(),
		Recipients: make(map[string][]byte, len(recipients)),
		Metadata:   make(map[string]string), // metadata is plaintext!
	}

	recipients = append(recipients, signKey.Fingerprint())
	sort.Strings(recipients)

	for _, recp := range recipients {
		// skip duplicates
		if _, found := hdr.Recipients[recp]; found {
			continue
		}

		r, err := x.encryptForRecipient(signKey, sk, recp)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encrypt session key for recipient %s: %s", recp, err)
		}

		hdr.Recipients[recp] = r
	}

	return hdr, nil
}

// encryptForRecipients encrypts the given session key for the given recipient
func (x *XC) encryptForRecipient(sender *keyring.PrivateKey, sk []byte, recipient string) ([]byte, error) {
	recp := x.pubring.Get(recipient)
	if recp == nil {
		return nil, fmt.Errorf("recipient public key not available for %s", recipient)
	}

	var recipientPublicKey [32]byte
	copy(recipientPublicKey[:], recp.PublicKey[:])

	// unlock sender key
	if err := x.decryptPrivateKey(sender); err != nil {
		return nil, err
	}

	// we need to copy the byte silces to byte arrays for box
	var senderPrivateKey [32]byte
	pk := sender.PrivateKey()
	copy(senderPrivateKey[:], pk[:])

	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	return box.Seal(nonce[:], sk, &nonce, &recipientPublicKey, &senderPrivateKey), nil
}

// encryptBody generates a random session key and a nonce and encrypts the given
// plaintext with those. it returns all three
func encryptBody(plaintext []byte) ([]byte, []*xcpb.Chunk, error) {
	// generate session / encryption key
	sessionKey := make([]byte, 32)
	if _, err := crypto_rand.Read(sessionKey); err != nil {
		return nil, nil, err
	}

	chunks := make([]*xcpb.Chunk, 0, (len(plaintext)/chunkSizeMax)+1)
	offset := 0

	for offset < len(plaintext) {
		// use a sequential nonce to prevent chunk reordering.
		// since the pair of key and nonce has to be unique and we're
		// generating a new random key for each message, this is OK
		nonce := make([]byte, 12)
		binary.BigEndian.PutUint64(nonce, uint64(len(chunks)))

		// initialize the AEAD with the generated session key
		cp, err := chacha20poly1305.New(sessionKey)
		if err != nil {
			return nil, nil, err
		}

		// encrypt the plaintext using the random nonce
		ciphertext := cp.Seal(nil, nonce, plaintext[offset:min(len(plaintext), offset+chunkSizeMax)], nil)
		chunks = append(chunks, &xcpb.Chunk{
			Body: ciphertext,
		})
		offset += chunkSizeMax
	}

	return sessionKey, chunks, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
