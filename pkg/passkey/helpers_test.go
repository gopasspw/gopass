package passkey

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unhex decodes a hex string.
func unhex(tb testing.TB, s string) []byte {
	tb.Helper()

	buf, err := hex.DecodeString(s)
	require.NoError(tb, err)

	return buf
}

// vectorKey returns the private key of a test vector of the specification.
func vectorKey(tb testing.TB, v specVector) crypto.Signer {
	tb.Helper()

	switch v.coseKey {
	case specES256.coseKey:
		key, err := ecdsa.ParseRawPrivateKey(elliptic.P256(), unhex(tb, v.privateKey))
		require.NoError(tb, err)

		return key
	case specEdDSA.coseKey:
		return ed25519.NewKeyFromSeed(unhex(tb, v.privateKey))
	default:
		return specRSAKey(tb)
	}
}

// specRSAKey returns the RSA key of the test vectors. It is "constructed
// from the two smallest Mersenne primes 2^p - 1 such that p >= 1024".
func specRSAKey(tb testing.TB) *rsa.PrivateKey {
	tb.Helper()

	one := big.NewInt(1)
	p := new(big.Int).Sub(new(big.Int).Lsh(one, 1279), one)
	q := new(big.Int).Sub(new(big.Int).Lsh(one, 2203), one)

	e := big.NewInt(65537)
	phi := new(big.Int).Mul(new(big.Int).Sub(p, one), new(big.Int).Sub(q, one))
	d := new(big.Int).ModInverse(e, phi)
	require.NotNil(tb, d)

	key := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{N: new(big.Int).Mul(p, q), E: int(e.Int64())},
		D:         d,
		Primes:    []*big.Int{p, q},
	}
	key.Precompute()

	return key
}

// vectorCredential returns the credential of a test vector.
func vectorCredential(tb testing.TB, v specVector, backup bool) *Credential {
	tb.Helper()

	return &Credential{
		ID:             unhex(tb, v.credentialID),
		RPID:           "example.org",
		UserHandle:     []byte("user"),
		Key:            vectorKey(tb, v),
		BackupEligible: backup,
		BackupState:    backup,
	}
}

// verify checks a WebAuthn signature the way a Relying Party does.
func verify(tb testing.TB, pub crypto.PublicKey, data, sig []byte) bool {
	tb.Helper()

	digest := sha256.Sum256(data)

	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		return ecdsa.VerifyASN1(k, digest[:], sig)
	case ed25519.PublicKey:
		return ed25519.Verify(k, data, sig)
	case *rsa.PublicKey:
		return rsa.VerifyPKCS1v15(k, crypto.SHA256, digest[:], sig) == nil
	default:
		tb.Fatalf("unsupported public key %T", pub)

		return false
	}
}

// verifyAssertion checks an assertion the way a Relying Party does, given
// the public key it stored at registration as a SubjectPublicKeyInfo.
func verifyAssertion(tb testing.TB, spki []byte, rsp *AuthenticationResponse) bool {
	tb.Helper()

	pub, err := x509.ParsePKIXPublicKey(spki)
	require.NoError(tb, err)

	hash := sha256.Sum256(rsp.Response.ClientDataJSON)

	return verify(tb, pub, append(append([]byte{}, rsp.Response.AuthenticatorData...), hash[:]...), rsp.Response.Signature)
}

// parsedAuthData is a decoded authenticator data structure.
type parsedAuthData struct {
	rpIDHash     []byte
	flags        Flags
	signCount    uint32
	aaguid       []byte
	credentialID []byte
	coseKey      []byte
}

// parseAuthData decodes authenticator data.
func parseAuthData(tb testing.TB, data []byte) parsedAuthData {
	tb.Helper()

	require.GreaterOrEqual(tb, len(data), 37)

	p := parsedAuthData{
		rpIDHash:  data[:32],
		flags:     Flags(data[32]),
		signCount: binary.BigEndian.Uint32(data[33:37]),
	}

	rest := data[37:]
	if !p.flags.Has(FlagAttestedCredentialData) {
		require.Empty(tb, rest)

		return p
	}

	require.GreaterOrEqual(tb, len(rest), 18)
	p.aaguid = rest[:16]
	n := int(binary.BigEndian.Uint16(rest[16:18]))
	require.GreaterOrEqual(tb, len(rest), 18+n)
	p.credentialID = rest[18 : 18+n]
	p.coseKey = rest[18+n:]

	return p
}

// cborDecode decodes one data item of the subset of CBOR the encoder emits
// and returns the remaining input. Integers are returned as int64, maps as
// map[any]any.
func cborDecode(tb testing.TB, data []byte) (any, []byte) {
	tb.Helper()

	require.NotEmpty(tb, data)

	major, info := data[0]>>5, data[0]&0x1f
	data = data[1:]

	var arg uint64

	switch {
	case info < 24:
		arg = uint64(info)
	case info == 24:
		arg, data = uint64(data[0]), data[1:]
	case info == 25:
		arg, data = uint64(binary.BigEndian.Uint16(data)), data[2:]
	case info == 26:
		arg, data = uint64(binary.BigEndian.Uint32(data)), data[4:]
	default:
		tb.Fatalf("unsupported additional information %d", info)
	}

	switch major {
	case cborMajorUint:
		return int64(arg), data
	case cborMajorNeg:
		return -1 - int64(arg), data
	case cborMajorBytes:
		return data[:arg], data[arg:]
	case cborMajorText:
		return string(data[:arg]), data[arg:]
	case cborMajorMap:
		m := make(map[any]any, arg)
		for range arg {
			var k, v any
			k, data = cborDecode(tb, data)
			v, data = cborDecode(tb, data)
			m[k] = v
		}

		return m, data
	default:
		tb.Fatalf("unsupported major type %d", major)

		return nil, nil
	}
}

// testContext returns a context with an in-memory config, so that tests do
// not depend on the config of the user running them.
func testContext(tb testing.TB) context.Context {
	tb.Helper()

	return config.NewContextInMemory()
}

// errReadOnly is returned by readOnlyStore for every write.
var errReadOnly = errors.New("store is read-only")

// readOnlyStore is a store that can not be written to.
type readOnlyStore struct {
	SecretStore
}

func (readOnlyStore) Set(context.Context, string, gopass.Byter) error {
	return errReadOnly
}

// timeUntil returns the number of seconds until t.
func timeUntil(t time.Time) float64 {
	return time.Until(t).Seconds()
}

// p384PKCS8 returns a private key on a curve that is not supported.
func p384PKCS8(tb testing.TB) []byte {
	tb.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(tb, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(tb, err)

	return der
}

// must returns v, or panics if there is an error.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

// equalBytes asserts that got is exactly want. It is used where the
// serialization matters, not just the value that was serialized.
func equalBytes(tb testing.TB, want string, got []byte) {
	tb.Helper()

	assert.Equal(tb, want, string(got))
}

// errPush is returned by unsyncedStore for every write.
var errPush = errors.New("failed to push to git remote")

// unsyncedStore is a store that writes a secret and then reports a failure,
// like a store whose remote can not be reached.
type unsyncedStore struct {
	SecretStore
}

func (u unsyncedStore) Set(ctx context.Context, name string, sec gopass.Byter) error {
	if err := u.SecretStore.Set(ctx, name, sec); err != nil {
		return err
	}

	return errPush
}
