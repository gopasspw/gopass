package passkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCBORHead(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   int64
		want string
	}{
		{0, "00"},
		{23, "17"},
		{24, "1818"},
		{255, "18ff"},
		{256, "190100"},
		{65535, "19ffff"},
		{65536, "1a00010000"},
		{4294967296, "1b0000000100000000"},
		{-1, "20"},
		{-7, "26"},
		{-24, "37"},
		{-25, "3818"},
		{-257, "390100"},
	} {
		assert.Equal(t, tc.want, hex.EncodeToString(cborInt(tc.in)), "%d", tc.in)
	}

	assert.Equal(t, "40", hex.EncodeToString(cborBytes(nil)))
	assert.Equal(t, "4401020304", hex.EncodeToString(cborBytes([]byte{1, 2, 3, 4})))
	assert.Equal(t, "60", hex.EncodeToString(cborText("")))
	assert.Equal(t, "63666d74", hex.EncodeToString(cborText("fmt")))
	assert.Equal(t, "a0", hex.EncodeToString(cborMap()))
}

func TestCBORMapIsCanonical(t *testing.T) {
	t.Parallel()

	// Keys are sorted by major type, then by length, then bytewise, no
	// matter in which order they are passed in.
	got := cborMap(
		cborPair{cborText("authData"), cborInt(4)},
		cborPair{cborInt(-2), cborInt(3)},
		cborPair{cborText("fmt"), cborInt(5)},
		cborPair{cborInt(3), cborInt(1)},
		cborPair{cborInt(-1), cborInt(2)},
		cborPair{cborText("b"), cborInt(7)},
		cborPair{cborInt(1), cborInt(0)},
		cborPair{cborText("attStmt"), cborInt(6)},
		cborPair{cborText("a"), cborInt(8)},
		cborPair{cborInt(256), cborInt(9)},
	)

	want := "aa" +
		"0100" + "0301" + "19010009" + // unsigned integers, shortest first
		"2002" + "2103" + // negative integers
		"616108" + "616207" + "63666d7405" + "6761747453746d7406" + "68617574684461746104" // text

	assert.Equal(t, want, hex.EncodeToString(got))
}

func TestCOSEKeyVectors(t *testing.T) {
	t.Parallel()

	for name, v := range map[string]specVector{
		"ES256": specES256,
		"RS256": specRS256,
		"EdDSA": specEdDSA,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := coseKey(vectorKey(t, v))
			require.NoError(t, err)
			assert.Equal(t, v.coseKey, hex.EncodeToString(got))
		})
	}
}

func TestCOSEKeyHasNoOptionalParameters(t *testing.T) {
	t.Parallel()

	for _, alg := range SupportedAlgorithms() {
		key, err := generateKey(alg)
		require.NoError(t, err)

		buf, err := coseKey(key)
		require.NoError(t, err)

		decoded, rest := cborDecode(t, buf)
		require.Empty(t, rest)

		m, ok := decoded.(map[any]any)
		require.True(t, ok)

		// "kty" and "alg" plus the parameters required by the key type
		assert.Equal(t, int64(alg), m[int64(coseKeyAlg)], alg)

		switch alg {
		case ES256:
			assert.Len(t, m, 5)
			assert.Equal(t, int64(coseKtyEC2), m[int64(coseKeyKty)])
			assert.Len(t, m[int64(coseKeyP2)], 32)
			assert.Len(t, m[int64(coseKeyP3)], 32)
		case EdDSA:
			assert.Len(t, m, 4)
			assert.Equal(t, int64(coseKtyOKP), m[int64(coseKeyKty)])
			assert.Len(t, m[int64(coseKeyP2)], 32)
		case RS256:
			assert.Len(t, m, 4)
			assert.Equal(t, int64(coseKtyRSA), m[int64(coseKeyKty)])
			assert.Len(t, m[int64(coseKeyP1)], rsaKeyBits/8)
			assert.Equal(t, []byte{1, 0, 1}, m[int64(coseKeyP2)])
		}
	}
}

func TestSign(t *testing.T) {
	t.Parallel()

	data := []byte("authenticator data and client data hash")

	for _, alg := range SupportedAlgorithms() {
		t.Run(alg.String(), func(t *testing.T) {
			t.Parallel()

			key, err := generateKey(alg)
			require.NoError(t, err)

			got, err := algorithmOf(key)
			require.NoError(t, err)
			assert.Equal(t, alg, got)

			sig, err := sign(key, data)
			require.NoError(t, err)
			assert.True(t, verify(t, key.Public(), data, sig))
			assert.False(t, verify(t, key.Public(), append([]byte("x"), data...), sig))

			if alg == EdDSA {
				// a raw signature, not an ASN.1 structure
				assert.Len(t, sig, 64)
			}
		})
	}
}

func TestPrivateKeyRoundTrip(t *testing.T) {
	t.Parallel()

	for _, alg := range SupportedAlgorithms() {
		key, err := generateKey(alg)
		require.NoError(t, err)

		der, err := marshalPrivateKey(key)
		require.NoError(t, err)

		parsed, err := parsePrivateKey(der)
		require.NoError(t, err)

		want, err := coseKey(key)
		require.NoError(t, err)
		got, err := coseKey(parsed)
		require.NoError(t, err)
		assert.Equal(t, want, got, alg)
	}
}

func TestUnsupportedKeys(t *testing.T) {
	t.Parallel()

	p384, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	_, err = algorithmOf(p384)
	require.ErrorIs(t, err, ErrUnsupportedKey)
	_, err = coseKey(p384)
	require.ErrorIs(t, err, ErrUnsupportedKey)
	_, err = sign(p384, []byte("data"))
	require.ErrorIs(t, err, ErrUnsupportedKey)
	_, err = marshalPrivateKey(p384)
	require.ErrorIs(t, err, ErrUnsupportedKey)

	// keys of an unsupported size are not accepted either
	key := specRSAKey(t)
	_, err = algorithmOf(&rsa.PrivateKey{PublicKey: rsa.PublicKey{N: key.Primes[0], E: key.E}, D: key.D})
	require.ErrorIs(t, err, ErrUnsupportedKey)

	// In-memory keys have not gone through the parser's exponent checks.
	for _, e := range []int64{-1, 0, 1, 2, 65536, 1<<31 + 1} {
		if int64(int(e)) != e {
			continue // not representable on 32-bit platforms
		}
		malformed := &rsa.PrivateKey{PublicKey: rsa.PublicKey{N: key.N, E: int(e)}, D: key.D}
		_, err := coseKey(malformed)
		require.ErrorIs(t, err, ErrUnsupportedKey, "exponent %d", e)
	}

	_, err = parsePrivateKey([]byte("not a key"))
	require.Error(t, err)

	_, err = generateKey(Algorithm(-35))
	require.ErrorIs(t, err, ErrUnsupportedKey)
}

func TestParseAlgorithm(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]Algorithm{
		"ES256":   ES256,
		"es256":   ES256,
		" EdDSA ": EdDSA,
		"ed25519": EdDSA,
		"RS256":   RS256,
	} {
		got, err := ParseAlgorithm(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
		assert.True(t, got.Supported())
	}

	_, err := ParseAlgorithm("ES512")
	require.Error(t, err)

	assert.Equal(t, "ES256", ES256.String())
	assert.Equal(t, "EdDSA", EdDSA.String())
	assert.Equal(t, "RS256", RS256.String())
	assert.Equal(t, "COSE(-36)", Algorithm(-36).String())
	assert.False(t, Algorithm(-36).Supported())
}
