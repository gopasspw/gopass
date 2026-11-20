package passkey

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/publicsuffix"
)

// FuzzResolveRPID checks what must hold for every pair of an origin and an
// RP ID that is accepted, whatever they look like.
func FuzzResolveRPID(f *testing.F) {
	for _, seed := range []struct {
		origin string
		rpID   string
		set    bool
	}{
		{"https://example.com", "", false},
		{"https://login.example.com:8443", "example.com", true},
		{"https://a.b.example.co.uk", "example.co.uk", true},
		{"https://bücher.example", "", false},
		{"http://localhost:8000", "localhost", true},
		{"https://example.com", "com", true},
		{"https://example.com", "", true},
		{"https://example.com.evil.example", "example.com", true},
		{"https://example.com\\@evil.example", "example.com", true},
		{"https://127.1", "", false},
		{"https://[2001:db8::1]:8443", "", false},
		{"https://user:pass@example.com:65536/path?query#fragment", "example.com", true},
		{"null", "", false},
	} {
		f.Add(seed.origin, seed.rpID, seed.set)
	}

	f.Fuzz(func(t *testing.T, origin, rpID string, set bool) {
		var requested *string
		if set {
			requested = &rpID
		}

		got, err := ResolveRPID(origin, requested)
		if err != nil {
			assert.Empty(t, got)
			assert.Contains(t, []ErrorName{SecurityError, NotAllowedError}, NameOf(err))

			return
		}

		// The origin is a secure context, and it is serialized the way it
		// stays when it is parsed again.
		serialized, host, err := parseOrigin(origin)
		require.NoError(t, err)

		again, _, err := parseOrigin(serialized)
		require.NoError(t, err)
		assert.Equal(t, serialized, again)

		if !strings.HasPrefix(serialized, "https://") {
			assert.Regexp(t, `^http://localhost(:\d+)?$`, serialized)
		}

		// The RP ID is the one that was asked for, in its canonical form.
		domain, err := normalizeDomain(got)
		require.NoError(t, err)
		assert.Equal(t, domain, got)

		if set {
			assert.Equal(t, rpID, got)
		}

		// It is the host of the origin or a parent domain of it, and in
		// the latter case not a public suffix.
		if got == host {
			return
		}

		assert.True(t, strings.HasSuffix(host, "."+got), "%q is not a suffix of %q", got, host)

		suffix, _ := publicsuffix.PublicSuffix(got)
		assert.NotEqual(t, suffix, got, "%q is a public suffix", got)

		hostSuffix, _ := publicsuffix.PublicSuffix(host)
		assert.False(t, strings.HasSuffix("."+hostSuffix, "."+got), "%q is part of the public suffix of %q", got, host)
	})
}

// FuzzDecode checks that a stored credential is either rejected or usable.
func FuzzDecode(f *testing.F) {
	for _, alg := range SupportedAlgorithms() {
		cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "alice", Algorithm: alg})
		require.NoError(f, err)

		buf, err := cred.Encode()
		require.NoError(f, err)
		f.Add(buf)
	}

	f.Add([]byte(`{"v":1,"rpId":"example.com","userHandle":"aGFuZGxl","credentialId":"aWQ","key":"AAAA"}`))
	f.Add([]byte(`{"v":2}`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		cred, err := Decode(data)
		if err != nil {
			assert.Nil(t, cred)
			require.ErrorIs(t, err, ErrInvalidCredential)

			return
		}

		buf, err := cred.Encode()
		require.NoError(t, err)
		assert.NotContains(t, string(buf), "\n")

		again, err := Decode(buf)
		require.NoError(t, err)
		assert.Equal(t, cred.ID, again.ID)
		assert.Equal(t, cred.RPID, again.RPID)
		assert.Equal(t, cred.UserHandle, again.UserHandle)
		assert.Equal(t, cred.SignCount, again.SignCount)

		att, err := cred.Attest(Consent{Present: true})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(att.AuthenticatorData), 37+16+2+len(cred.ID))

		hash := sha256.Sum256(data)

		assertion, err := cred.Assert(hash[:], Consent{Present: true})
		require.NoError(t, err)
		assert.Len(t, assertion.AuthenticatorData, 37)
		assert.True(t, verify(t, cred.Key.Public(), append(assertion.AuthenticatorData, hash[:]...), assertion.Signature))
	})
}

// FuzzImportCXF checks that importing never yields a credential that can not
// be stored.
func FuzzImportCXF(f *testing.F) {
	cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "alice"})
	require.NoError(f, err)

	var doc bytes.Buffer

	_, err = ExportCXF(&doc, []Entry{{Name: "passkeys/example.com/alice", Credential: cred}})
	require.NoError(f, err)

	f.Add(doc.Bytes())
	f.Add([]byte(`{"version":{"major":1,"minor":0},"accounts":[{"items":[{"credentials":[{"type":"passkey"},{"type":"note"},42]}]}]}`))
	f.Add([]byte(`{"version":{"major":1},"accounts":null}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		creds, _ := ImportCXF(bytes.NewReader(data))

		for _, cred := range creds {
			require.NoError(t, cred.validate())
			assert.False(t, cred.Counting())

			_, err := cred.Encode()
			require.NoError(t, err)

			assert.Equal(t, sanitizeName(accountName(cred)), accountName(cred))
		}
	})
}

// FuzzClientDataJSON checks that the hand-written serialization of the
// client data is JSON that carries what was put in.
func FuzzClientDataJSON(f *testing.F) {
	f.Add("webauthn.get", []byte("challenge"), "https://example.com", false, "")
	f.Add("webauthn.create", []byte{0xfb, 0xff}, "https://example.com", true, "https://shop.example")
	f.Add("a\"b\\c\n\x00\x1f\x7f\u2028\U0001f600\xff", []byte(nil), "\\\"", true, "\u0000")

	f.Fuzz(func(t *testing.T, typ string, challenge []byte, origin string, crossOrigin bool, topOrigin string) {
		cd := ClientData{Type: typ, Challenge: challenge, Origin: origin, CrossOrigin: crossOrigin, TopOrigin: topOrigin}
		buf := cd.JSON()

		var got struct {
			Type        string    `json:"type"`
			Challenge   Base64URL `json:"challenge"`
			Origin      string    `json:"origin"`
			CrossOrigin *bool     `json:"crossOrigin"`
			TopOrigin   *string   `json:"topOrigin"`
		}
		require.NoError(t, json.Unmarshal(buf, &got), string(buf))

		// Every byte that is not valid UTF-8 is replaced.
		assert.Equal(t, string([]rune(typ)), got.Type)
		assert.Equal(t, string([]rune(origin)), got.Origin)
		assert.True(t, bytes.Equal(challenge, got.Challenge))

		require.NotNil(t, got.CrossOrigin)
		assert.Equal(t, crossOrigin, *got.CrossOrigin)
		assert.Equal(t, crossOrigin && topOrigin != "", got.TopOrigin != nil)

		// The members come in the order a verifier without a JSON parser
		// relies on.
		assert.True(t, bytes.HasPrefix(buf, []byte(`{"type":"`)))
		assert.Less(t, bytes.Index(buf, []byte(`,"challenge":"`)), bytes.LastIndex(buf, []byte(`,"origin":"`)))
		assert.Less(t, bytes.LastIndex(buf, []byte(`,"origin":"`)), bytes.LastIndex(buf, []byte(`,"crossOrigin":`)))
	})
}
