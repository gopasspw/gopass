// Package crypto implements the transport encryption algorithms defined by
// the freedesktop.org Secret Service specification.
//
// A [Session] encrypts and decrypts the payload of a D-Bus Secret struct for
// one client that called OpenSession. Two algorithms are defined by the spec:
//
//   - "plain" — no transport encryption. This is secure on the session bus,
//     which is a local UNIX socket with kernel-enforced access control.
//   - "dh-ietf1024-sha256-aes128-cbc-pkcs7" — Diffie-Hellman key exchange with
//     AES-128-CBC payload encryption. Required for compatibility with
//     libsecret-based clients.
//
// The package deliberately has no build constraint: the crypto is portable and
// is unit-tested on every platform. Only the D-Bus service that uses it is
// Linux-only.
package crypto

import "fmt"

// Session algorithm names as defined by the Secret Service specification.
const (
	// AlgorithmPlain performs no transport encryption.
	AlgorithmPlain = "plain"
	// AlgorithmDHAES is Diffie-Hellman with AES-128-CBC.
	AlgorithmDHAES = "dh-ietf1024-sha256-aes128-cbc-pkcs7"
)

// Session is an established transport-encryption session.
type Session interface {
	// Encrypt encrypts plaintext and returns the IV (Parameters) and the
	// ciphertext.
	Encrypt(plaintext []byte) (params, ciphertext []byte, err error)
	// Decrypt decrypts ciphertext using params as the IV.
	Decrypt(params, ciphertext []byte) (plaintext []byte, err error)
	// Close releases any key material held by the session.
	Close() error
}

// New negotiates a new Session for the given algorithm.
//
// For [AlgorithmPlain] clientInput must be empty and the returned output is
// empty. For [AlgorithmDHAES] clientInput is the client's DH public key and the
// returned output is the server's DH public key, both left-padded to 128 bytes.
func New(algorithm string, clientInput []byte) (Session, []byte, error) {
	switch algorithm {
	case AlgorithmPlain:
		return plainSession{}, nil, nil
	case AlgorithmDHAES:
		return newDHSession(clientInput)
	default:
		return nil, nil, fmt.Errorf("unsupported algorithm: %q", algorithm)
	}
}
