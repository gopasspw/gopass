package xc

import (
	"context"
	crypto_rand "crypto/rand"
	stdbin "encoding/binary"
	"fmt"
	"io"

	"github.com/alecthomas/binary"
	"github.com/justwatchcom/gopass/pkg/backend/crypto/xc/xcpb"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/secretbox"
)

// EncryptStream encrypts the plaintext using a slightly modified on disk-format
// suitable for streaming
func (x *XC) EncryptStream(ctx context.Context, plaintext io.Reader, recipients []string, ciphertext io.Writer) error {
	privKeyIDs := x.secring.KeyIDs()
	if len(privKeyIDs) < 1 {
		return fmt.Errorf("no signing keys available on our keyring")
	}
	privKey := x.secring.Get(privKeyIDs[0])

	// generate session / encryption key
	var sessionKey [32]byte
	if _, err := crypto_rand.Read(sessionKey[:]); err != nil {
		return err
	}

	// encrypt the session key per recipient
	header, err := x.encryptHeader(ctx, privKey, sessionKey[:], recipients)
	if err != nil {
		return errors.Wrapf(err, "failed to encrypt header: %s", err)
	}

	// create the encoder
	enc := binary.NewEncoder(ciphertext)

	// write verion
	if err := enc.Encode(0x1); err != nil {
		return err
	}
	// write header
	if err := enc.Encode(header); err != nil {
		return err
	}
	// write body
	num := 0
	buf := make([]byte, chunkSizeMax)
	encbuf := make([]byte, 8)
	for {
		n, err := plaintext.Read(buf)
		if err := x.encryptChunk(sessionKey, num, buf[:n], encbuf, ciphertext); err != nil {
			return err
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		num++
	}
}

func (x *XC) encryptChunk(sessionKey [32]byte, num int, buf, encbuf []byte, ciphertext io.Writer) error {
	// use a sequential nonce to prevent chunk reordering.
	// since the pair of key and nonce has to be unique and we're
	// generating a new random key for each message, this is OK
	var nonce [24]byte
	binary.BigEndian.PutUint64(nonce[:], uint64(num))

	// encrypt the plaintext using the random nonce
	cipherBlock := secretbox.Seal(nil, buf, &nonce, &sessionKey)

	// write ciphertext block length
	l := stdbin.PutUvarint(encbuf, uint64(len(cipherBlock)))
	if _, err := ciphertext.Write(encbuf[:l]); err != nil {
		return err
	}

	// write ciphertext block data
	_, err := ciphertext.Write(cipherBlock)
	return err
}

// DecryptStream decrypts an stream encrypted with EncryptStream
func (x *XC) DecryptStream(ctx context.Context, ciphertext io.Reader, plaintext io.Writer) error {
	dec := binary.NewDecoder(ciphertext)

	// read version
	ver := 0
	if err := dec.Decode(&ver); err != nil {
		return err
	}
	if ver != 0x1 {
		return fmt.Errorf("wrong version")
	}
	// read header
	header := &xcpb.Header{}
	if err := dec.Decode(header); err != nil {
		return err
	}

	// try to find a suiteable decryption key in the header
	sk, err := x.decryptSessionKey(ctx, header)
	if err != nil {
		return err
	}

	var secretKey [32]byte
	copy(secretKey[:], sk)

	// read body
	num := 0
	var buf []byte
	br := &byteReader{ciphertext}
	for {
		l, err := stdbin.ReadUvarint(br)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		buf = make([]byte, l)
		n, err := br.Read(buf)
		if err := x.decryptChunk(secretKey, num, buf[:n], plaintext); err != nil {
			return err
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		num++
	}
}

func (x *XC) decryptChunk(secretKey [32]byte, num int, buf []byte, plaintext io.Writer) error {
	// reconstruct nonce from chunk number
	// in case chunks have been reordered by some adversary
	// decryption will fail
	var nonce [24]byte
	binary.BigEndian.PutUint64(nonce[:], uint64(num))

	// decrypt and verify the ciphertext
	plain, ok := secretbox.Open(nil, buf, &nonce, &secretKey)
	if !ok {
		return fmt.Errorf("failed to decrypt chunk %d", num)
	}

	_, err := plaintext.Write(plain)
	return err
}

type byteReader struct {
	io.Reader
}

func (b *byteReader) ReadByte() (byte, error) {
	var buf [1]byte
	if _, err := io.ReadFull(b, buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}
