package passkey

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"math/big"
	"strings"
)

// Algorithm is a COSEAlgorithmIdentifier, a number from the IANA COSE
// Algorithms registry that identifies a signature algorithm.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-alg-identifier
type Algorithm int

const (
	// ES256 is ECDSA using the P-256 curve and SHA-256.
	ES256 Algorithm = -7
	// EdDSA is the Edwards-curve Digital Signature Algorithm. Only the
	// Ed25519 curve is supported.
	EdDSA Algorithm = -8
	// RS256 is RSASSA-PKCS1-v1_5 using SHA-256.
	RS256 Algorithm = -257
)

const (
	// rsaKeyBits is the size of generated RSA keys.
	rsaKeyBits = 2048
	// rsaMinKeyBits is the size below which RSA keys are rejected.
	rsaMinKeyBits = 2048
)

// COSE key types, parameters and curves, see RFC 9052 and RFC 9053.
const (
	coseKeyKty = 1
	coseKeyAlg = 3
	// The meaning of the negative labels depends on the key type.
	coseKeyP1 = -1
	coseKeyP2 = -2
	coseKeyP3 = -3

	coseKtyOKP = 1
	coseKtyEC2 = 2
	coseKtyRSA = 3

	coseCrvP256    = 1
	coseCrvEd25519 = 6
)

// String returns the name of the algorithm.
func (a Algorithm) String() string {
	switch a {
	case ES256:
		return "ES256"
	case EdDSA:
		return "EdDSA"
	case RS256:
		return "RS256"
	default:
		return fmt.Sprintf("COSE(%d)", int(a))
	}
}

// Supported reports whether credentials can be created for the algorithm.
func (a Algorithm) Supported() bool {
	switch a {
	case ES256, EdDSA, RS256:
		return true
	default:
		return false
	}
}

// SupportedAlgorithms returns the algorithms credentials can be created for,
// in the order of preference.
func SupportedAlgorithms() []Algorithm {
	return []Algorithm{ES256, EdDSA, RS256}
}

// ParseAlgorithm returns the algorithm for a name like "ES256". The name is
// not case sensitive.
func ParseAlgorithm(name string) (Algorithm, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "es256", "p256", "p-256":
		return ES256, nil
	case "eddsa", "ed25519":
		return EdDSA, nil
	case "rs256", "rsa":
		return RS256, nil
	default:
		return 0, fmt.Errorf("unknown algorithm %q, use one of ES256, EdDSA or RS256", name)
	}
}

// generateKey creates a new private key for the algorithm.
func generateKey(alg Algorithm) (crypto.Signer, error) {
	switch alg {
	case ES256:
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
		}

		return key, nil
	case EdDSA:
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}

		return key, nil
	case RS256:
		key, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}

		return key, nil
	default:
		return nil, fmt.Errorf("%w: algorithm %s", ErrUnsupportedKey, alg)
	}
}

// algorithmOf returns the algorithm a private key is used with. It rejects
// keys that can not be used by this package.
func algorithmOf(key crypto.Signer) (Algorithm, error) {
	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		if k == nil || k.Curve != elliptic.P256() {
			return 0, fmt.Errorf("%w: only the P-256 curve is supported for ECDSA", ErrUnsupportedKey)
		}

		// A key that was not generated or parsed can lack its values.
		if k.D == nil || k.X == nil || k.Y == nil {
			return 0, fmt.Errorf("%w: malformed ECDSA key", ErrUnsupportedKey)
		}

		return ES256, nil
	case ed25519.PrivateKey:
		if len(k) != ed25519.PrivateKeySize {
			return 0, fmt.Errorf("%w: malformed Ed25519 key", ErrUnsupportedKey)
		}

		return EdDSA, nil
	case *rsa.PrivateKey:
		if k == nil || k.N == nil || k.D == nil || k.N.BitLen() < rsaMinKeyBits {
			return 0, fmt.Errorf("%w: RSA keys must have at least %d bits", ErrUnsupportedKey, rsaMinKeyBits)
		}

		return RS256, nil
	default:
		return 0, fmt.Errorf("%w: key type %T", ErrUnsupportedKey, key)
	}
}

// parsePrivateKey parses a PKCS#8 encoded private key.
func parsePrivateKey(der []byte) (crypto.Signer, error) {
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
	}

	key, ok := parsed.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("%w: key type %T", ErrUnsupportedKey, parsed)
	}

	if _, err := algorithmOf(key); err != nil {
		return nil, err
	}

	return key, nil
}

// marshalPrivateKey encodes a private key as PKCS#8.
func marshalPrivateKey(key crypto.Signer) ([]byte, error) {
	if _, err := algorithmOf(key); err != nil {
		return nil, err
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	return der, nil
}

// sign produces a signature over data in the format WebAuthn mandates for
// the algorithm of the key.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-signature-attestation-types
func sign(key crypto.Signer, data []byte) ([]byte, error) {
	alg, err := algorithmOf(key)
	if err != nil {
		return nil, err
	}

	var sig []byte

	switch alg {
	case ES256, RS256:
		// *ecdsa.PrivateKey returns an ASN.1 DER Ecdsa-Sig-Value and
		// *rsa.PrivateKey a raw PKCS #1 v1.5 signature. Both sign a digest.
		digest := sha256.Sum256(data)
		sig, err = key.Sign(rand.Reader, digest[:], crypto.SHA256)
	case EdDSA:
		// Ed25519 signs the message itself and returns the raw signature.
		sig, err = key.Sign(rand.Reader, data, crypto.Hash(0))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return sig, nil
}

// coseKey encodes the public key of a credential as a COSE_Key in the
// canonical CBOR form. It contains the required parameters and "alg", but no
// other optional parameters.
//
// See https://www.w3.org/TR/webauthn-3/#sctn-encoded-credPubKey-examples
func coseKey(key crypto.Signer) ([]byte, error) {
	alg, err := algorithmOf(key)
	if err != nil {
		return nil, err
	}

	switch pub := key.Public().(type) {
	case *ecdsa.PublicKey:
		// The coordinates are fixed length, big endian integers. The
		// uncompressed point is 0x04 || X || Y.
		point, err := pub.Bytes()
		if err != nil {
			return nil, fmt.Errorf("failed to encode ECDSA public key: %w", err)
		}
		size := (len(point) - 1) / 2

		return cborMap(
			cborPair{cborInt(coseKeyKty), cborInt(coseKtyEC2)},
			cborPair{cborInt(coseKeyAlg), cborInt(int64(alg))},
			cborPair{cborInt(coseKeyP1), cborInt(coseCrvP256)},
			cborPair{cborInt(coseKeyP2), cborBytes(point[1 : 1+size])},
			cborPair{cborInt(coseKeyP3), cborBytes(point[1+size:])},
		), nil
	case ed25519.PublicKey:
		return cborMap(
			cborPair{cborInt(coseKeyKty), cborInt(coseKtyOKP)},
			cborPair{cborInt(coseKeyAlg), cborInt(int64(alg))},
			cborPair{cborInt(coseKeyP1), cborInt(coseCrvEd25519)},
			cborPair{cborInt(coseKeyP2), cborBytes(pub)},
		), nil
	case *rsa.PublicKey:
		return cborMap(
			cborPair{cborInt(coseKeyKty), cborInt(coseKtyRSA)},
			cborPair{cborInt(coseKeyAlg), cborInt(int64(alg))},
			cborPair{cborInt(coseKeyP1), cborBytes(pub.N.Bytes())},
			cborPair{cborInt(coseKeyP2), cborBytes(big.NewInt(int64(pub.E)).Bytes())},
		), nil
	default:
		return nil, fmt.Errorf("%w: public key type %T", ErrUnsupportedKey, pub)
	}
}
