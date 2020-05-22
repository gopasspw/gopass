package xc

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sort"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/gopasspw/gopass/internal/backend/crypto/xc/keyring"
	"github.com/gopasspw/gopass/internal/backend/crypto/xc/xcpb"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	// OnDiskVersion is the version of our on-disk format
	OnDiskVersion = 1
	// chunkSizeMax is chosen according to the NaCl recommendation:
	// "If in doubt, 16KB is a reasonable chunk size."
	// https://godoc.org/golang.org/x/crypto/nacl/secretbox
	chunkSizeMax = 16 * 1024
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
	header, err := x.encryptHeader(ctx, privKey, sk, recipients)
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
func (x *XC) encryptHeader(ctx context.Context, signKey *keyring.PrivateKey, sk []byte, recipients []string) (*xcpb.Header, error) {
	hdr := &xcpb.Header{
		Sender:     signKey.Fingerprint(),
		Recipients: make(map[string][]byte, len(recipients)),
	}

	recipients = append(recipients, signKey.Fingerprint())
	sort.Strings(recipients)

	for _, recp := range recipients {
		// skip duplicates
		if _, found := hdr.Recipients[recp]; found {
			continue
		}

		r, err := x.encryptForRecipient(ctx, signKey, sk, recp)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encrypt session key for recipient %s: %s", recp, err)
		}

		hdr.Recipients[recp] = r
	}

	return hdr, nil
}

// encryptForRecipients encrypts the given session key for the given recipient
func (x *XC) encryptForRecipient(ctx context.Context, sender *keyring.PrivateKey, sk []byte, recipient string) ([]byte, error) {
	recp := x.pubring.Get(recipient)
	if recp == nil {
		return nil, fmt.Errorf("recipient public key not available for %s", recipient)
	}

	var recipientPublicKey [32]byte
	copy(recipientPublicKey[:], recp.PublicKey[:])

	// unlock sender key
	if err := x.decryptPrivateKey(ctx, sender); err != nil {
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
	var sessionKey [32]byte
	if _, err := crypto_rand.Read(sessionKey[:]); err != nil {
		return nil, nil, err
	}

	chunks := make([]*xcpb.Chunk, 0, (len(plaintext)/chunkSizeMax)+1)
	offset := 0

	for offset < len(plaintext) {
		// use a sequential nonce to prevent chunk reordering.
		// since the pair of key and nonce has to be unique and we're
		// generating a new random key for each message, this is OK
		var nonce [24]byte
		binary.BigEndian.PutUint64(nonce[:], uint64(len(chunks)))

		// encrypt the plaintext using the random nonce
		chunks = append(chunks, &xcpb.Chunk{
			Body: secretbox.Seal(nil, plaintext[offset:min(len(plaintext), offset+chunkSizeMax)], &nonce, &sessionKey),
		})
		offset += chunkSizeMax
	}

	return sessionKey[:], chunks, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
