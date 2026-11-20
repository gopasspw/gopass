package passkey

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientDataJSONVectors(t *testing.T) {
	t.Parallel()

	for name, v := range map[string]specVector{
		"ES256": specES256,
		"RS256": specRS256,
		"EdDSA": specEdDSA,
	} {
		cd := ClientData{
			Type:      clientDataGet,
			Challenge: unhex(t, v.authChallenge),
			Origin:    "https://example.org",
		}
		assert.Equal(t, v.authClientData, hex.EncodeToString(cd.JSON()), name)

		hash := sha256.Sum256(unhex(t, v.authClientData))
		assert.Equal(t, hash[:], cd.Hash(), name)
	}

	// The registration of this vector has no additional members either.
	cd := ClientData{
		Type:      clientDataCreate,
		Challenge: unhex(t, specRS256.regChallenge),
		Origin:    "https://example.org",
	}
	assert.Equal(t, specRS256.regClientData, hex.EncodeToString(cd.JSON()))
}

func TestClientDataJSON(t *testing.T) {
	t.Parallel()

	cd := ClientData{
		Type:        clientDataGet,
		Challenge:   []byte{0xfb, 0xff},
		Origin:      "https://example.org",
		CrossOrigin: true,
		TopOrigin:   "https://example.com",
	}
	equalBytes(t, `{"type":"webauthn.get","challenge":"-_8","origin":"https://example.org","crossOrigin":true,"topOrigin":"https://example.com"}`, cd.JSON())

	// The top-level origin is only present for cross-origin requests.
	cd.CrossOrigin = false
	equalBytes(t, `{"type":"webauthn.get","challenge":"-_8","origin":"https://example.org","crossOrigin":false}`, cd.JSON())

	// Strings are escaped as mandated by CCDToString.
	cd = ClientData{Type: "a\"b\\c\n\x1f/<é\u2028"}
	want := `{"type":"a\"b\\c\u000a\u001f/<é` + "\u2028" + `","challenge":"","origin":"","crossOrigin":false}`
	equalBytes(t, want, cd.JSON())

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(cd.JSON(), &decoded))
	assert.Equal(t, cd.Type, decoded["type"])
}

func TestAttestVector(t *testing.T) {
	// not parallel: it replaces the AAGUID
	v := specES256

	aaguid := AAGUID
	t.Cleanup(func() { AAGUID = aaguid })
	copy(AAGUID[:], unhex(t, v.aaguid))

	cred := vectorCredential(t, v, true)

	att, err := cred.Attest(Consent{Present: true})
	require.NoError(t, err)

	assert.Equal(t, v.regAuthData, hex.EncodeToString(att.AuthenticatorData))
	assert.Equal(t, v.attestationObject, hex.EncodeToString(att.Object))
}

func TestAttest(t *testing.T) {
	t.Parallel()

	for _, alg := range SupportedAlgorithms() {
		cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), Algorithm: alg})
		require.NoError(t, err)

		att, err := cred.Attest(Consent{Present: true, Verified: true})
		require.NoError(t, err)

		decoded, rest := cborDecode(t, att.Object)
		require.Empty(t, rest)

		obj, ok := decoded.(map[any]any)
		require.True(t, ok)
		assert.Len(t, obj, 3)
		assert.Equal(t, "none", obj["fmt"])
		assert.Empty(t, obj["attStmt"])
		assert.Equal(t, att.AuthenticatorData, obj["authData"])

		ad := parseAuthData(t, att.AuthenticatorData)
		hash := sha256.Sum256([]byte("example.com"))
		assert.Equal(t, hash[:], ad.rpIDHash)
		assert.Equal(t, FlagUserPresent|FlagUserVerified|FlagBackupEligible|FlagBackupState|FlagAttestedCredentialData, ad.flags)
		assert.Zero(t, ad.signCount)
		assert.Equal(t, AAGUID[:], ad.aaguid)
		assert.Equal(t, cred.ID, ad.credentialID)

		pub, err := coseKey(cred.Key)
		require.NoError(t, err)
		assert.Equal(t, pub, ad.coseKey)
	}
}

func TestAssertVectors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		vector specVector
		backup bool
		// deterministic is true if the algorithm always produces the same
		// signature for the same input
		deterministic bool
	}{
		"ES256": {specES256, true, false},
		"RS256": {specRS256, true, true},
		"EdDSA": {specEdDSA, false, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := tc.vector
			cred := vectorCredential(t, v, tc.backup)
			hash := sha256.Sum256(unhex(t, v.authClientData))

			// the signature of the specification is valid for our key
			signed := append(unhex(t, v.authData), hash[:]...)
			require.True(t, verify(t, cred.Key.Public(), signed, unhex(t, v.signature)))

			got, err := cred.Assert(hash[:], Consent{Present: true})
			require.NoError(t, err)
			assert.Equal(t, v.authData, hex.EncodeToString(got.AuthenticatorData))
			assert.True(t, verify(t, cred.Key.Public(), signed, got.Signature))

			if tc.deterministic {
				assert.Equal(t, v.signature, hex.EncodeToString(got.Signature))
			}
		})
	}
}

func TestAssert(t *testing.T) {
	t.Parallel()

	cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle")})
	require.NoError(t, err)

	hash := sha256.Sum256([]byte("client data"))

	_, err = cred.Assert([]byte("too short"), Consent{Present: true})
	require.Error(t, err)

	// nothing is attested about the user that did not happen
	got, err := cred.Assert(hash[:], Consent{})
	require.NoError(t, err)
	assert.Equal(t, FlagBackupEligible|FlagBackupState, parseAuthData(t, got.AuthenticatorData).flags)

	got, err = cred.Assert(hash[:], Consent{Present: true, Verified: true})
	require.NoError(t, err)

	ad := parseAuthData(t, got.AuthenticatorData)
	assert.Equal(t, FlagUserPresent|FlagUserVerified|FlagBackupEligible|FlagBackupState, ad.flags)
	assert.Len(t, got.AuthenticatorData, 37)

	// a credential that is not backup eligible can not be backed up
	cred.BackupEligible = false
	got, err = cred.Assert(hash[:], Consent{Present: true})
	require.NoError(t, err)
	assert.Equal(t, FlagUserPresent, parseAuthData(t, got.AuthenticatorData).flags)
}

func TestSignCounter(t *testing.T) {
	t.Parallel()

	hash := sha256.Sum256([]byte("client data"))

	// synced credentials do not count
	cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle")})
	require.NoError(t, err)
	assert.False(t, cred.Counting())

	for range 3 {
		require.NoError(t, cred.Advance())

		got, err := cred.Assert(hash[:], Consent{Present: true})
		require.NoError(t, err)
		assert.Zero(t, parseAuthData(t, got.AuthenticatorData).signCount)
		assert.Zero(t, cred.SignCount)
	}

	// unless they are opted in
	cred, err = New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), SignCounter: true})
	require.NoError(t, err)
	assert.True(t, cred.Counting())

	att, err := cred.Attest(Consent{Present: true})
	require.NoError(t, err)

	last := parseAuthData(t, att.AuthenticatorData).signCount
	assert.Equal(t, uint32(1), last)

	for range 3 {
		require.NoError(t, cred.Advance())

		got, err := cred.Assert(hash[:], Consent{Present: true})
		require.NoError(t, err)

		count := parseAuthData(t, got.AuthenticatorData).signCount
		assert.Greater(t, count, last)
		assert.Equal(t, count, cred.SignCount)
		last = count
	}

	// Signing does not change the counter.
	_, err = cred.Assert(hash[:], Consent{Present: true})
	require.NoError(t, err)
	assert.Equal(t, last, cred.SignCount)

	cred.SignCount = ^uint32(0)
	require.ErrorIs(t, cred.Advance(), ErrInvalidCredential)
	assert.Equal(t, ^uint32(0), cred.SignCount)
}

func TestNew(t *testing.T) {
	t.Parallel()

	_, err := New(NewOptions{UserHandle: []byte("handle")})
	require.ErrorIs(t, err, ErrInvalidCredential)

	_, err = New(NewOptions{RPID: "example.com"})
	require.ErrorIs(t, err, ErrInvalidCredential)

	_, err = New(NewOptions{RPID: "example.com", UserHandle: make([]byte, 65)})
	require.ErrorIs(t, err, ErrInvalidCredential)

	_, err = New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), Algorithm: Algorithm(-36)})
	require.ErrorIs(t, err, ErrUnsupportedKey)

	// The RP ID is hashed as it is: only the canonical form is accepted.
	for _, rpID := range []string{"Example.com", "example.com.", "bücher.example", "192.0.2.1", "example.com/x", " example.com"} {
		_, err = New(NewOptions{RPID: rpID, UserHandle: []byte("handle")})
		require.ErrorIs(t, err, ErrInvalidCredential, rpID)
	}

	a, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "alice", UserDisplayName: "Alice"})
	require.NoError(t, err)

	b, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle")})
	require.NoError(t, err)

	assert.Len(t, a.ID, 32)
	assert.NotEqual(t, a.ID, b.ID)
	assert.Equal(t, ES256, a.Algorithm())
	assert.True(t, a.BackupEligible)
	assert.True(t, a.BackupState)
	assert.WithinDuration(t, time.Now(), a.Created, time.Minute)

	assert.Equal(t, "alice (Alice)", a.Label())
	assert.Equal(t, Base64URL(b.ID).String(), b.Label())

	b.UserDisplayName = "Bob"
	assert.Equal(t, "Bob", b.Label())
	b.UserName = "Bob"
	assert.Equal(t, "Bob", b.Label())

	// The names are chosen by a web site. The label is safe to print.
	b.UserName = "bob\x1b[31m\nSECRET\tRP ID\r\u202e\x00  x"
	b.UserDisplayName = strings.Repeat("\u00e4", 100)
	assert.Equal(t, "bob [31m SECRET RP ID x ("+strings.Repeat("\u00e4", 64)+"\u2026)", b.Label())
	assert.Contains(t, b.String(), "bob [31m SECRET RP ID x")

	b.UserName, b.UserDisplayName = "\x00\n", " \t "
	assert.Equal(t, Base64URL(b.ID).String(), b.Label())

	pa, err := a.PublicKey()
	require.NoError(t, err)
	pb, err := b.PublicKey()
	require.NoError(t, err)
	assert.NotEqual(t, pa, pb)
}

func TestCredentialEncoding(t *testing.T) {
	t.Parallel()

	for _, alg := range SupportedAlgorithms() {
		cred, err := New(NewOptions{
			RPID:            "example.com",
			UserHandle:      []byte{0xfb, 0xff, 0x00},
			UserName:        "alice@example.com",
			UserDisplayName: `Alice & "Bob": <admin>`,
			Algorithm:       alg,
		})
		require.NoError(t, err)

		buf, err := cred.Encode()
		require.NoError(t, err)

		// a single line that survives as the value of a key-value pair
		assert.NotContains(t, string(buf), "\n")
		assert.True(t, strings.HasPrefix(string(buf), `{"v":1,"rpId":"example.com","userName":"alice@example.com",`), string(buf))
		assert.Contains(t, string(buf), `"userHandle":"-_8A"`)
		assert.Contains(t, string(buf), `"signCount":0,`)
		assert.Contains(t, string(buf), `"backupEligible":true,"backupState":true`)
		assert.Contains(t, string(buf), `Alice & \"Bob\": <admin>`)

		got, err := Decode(buf)
		require.NoError(t, err)
		assert.Equal(t, cred.ID, got.ID)
		assert.Equal(t, cred.RPID, got.RPID)
		assert.Equal(t, cred.UserHandle, got.UserHandle)
		assert.Equal(t, cred.UserName, got.UserName)
		assert.Equal(t, cred.UserDisplayName, got.UserDisplayName)
		assert.Equal(t, cred.SignCount, got.SignCount)
		assert.Equal(t, cred.BackupEligible, got.BackupEligible)
		assert.Equal(t, cred.BackupState, got.BackupState)
		assert.True(t, cred.Created.Equal(got.Created))
		assert.Equal(t, alg, got.Algorithm())

		want, err := cred.PublicKey()
		require.NoError(t, err)
		pub, err := got.PublicKey()
		require.NoError(t, err)
		assert.Equal(t, want, pub)
	}
}

func TestCredentialDoesNotLeakItsKey(t *testing.T) {
	t.Parallel()

	for _, alg := range SupportedAlgorithms() {
		cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "alice", Algorithm: alg})
		require.NoError(t, err)

		stored, err := cred.Encode()
		require.NoError(t, err)

		var fields map[string]any
		require.NoError(t, json.Unmarshal(stored, &fields))

		key, ok := fields["key"].(string)
		require.True(t, ok)
		require.NotEmpty(t, key)

		// Whatever is derived from a credential for display or logging does
		// not contain the private key. Only Encode does.
		entry := Entry{Name: "passkeys/example.com/alice", Credential: cred}

		for name, out := range map[string]string{
			"json":       string(must(json.Marshal(cred))),
			"json entry": string(must(json.Marshal([]Entry{entry}))),
			"string":     cred.String(),
			"%v":         fmt.Sprintf("%v", cred),
			"%+v":        fmt.Sprintf("%+v", cred),
			"%+v entry":  fmt.Sprintf("%+v", entry),
		} {
			assert.NotContains(t, out, key, name)
			assert.NotContains(t, out, key[:16], name)
			assert.Contains(t, out, Base64URL(cred.ID).String(), name)
		}

		assert.JSONEq(t, fmt.Sprintf(
			`{"id":%q,"rpId":"example.com","userName":"alice","userDisplayName":"","algorithm":%d,"created":%q}`,
			Base64URL(cred.ID).String(), alg, cred.Created.Format(time.RFC3339),
		), string(must(json.Marshal(cred))))
	}

	assert.Equal(t, "passkey(nil)", (*Credential)(nil).String())
}

func TestCredentialDecodeInvalid(t *testing.T) {
	t.Parallel()

	cred, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle")})
	require.NoError(t, err)

	buf, err := cred.Encode()
	require.NoError(t, err)

	var fields map[string]any
	require.NoError(t, json.Unmarshal(buf, &fields))

	mutate := func(fn func(m map[string]any)) []byte {
		m := map[string]any{}
		for k, v := range fields {
			m[k] = v
		}
		fn(m)

		out, err := json.Marshal(m)
		require.NoError(t, err)

		return out
	}

	for name, in := range map[string][]byte{
		"not json":         []byte("passkey"),
		"empty":            nil,
		"future version":   mutate(func(m map[string]any) { m["v"] = 2 }),
		"no version":       mutate(func(m map[string]any) { delete(m, "v") }),
		"no key":           mutate(func(m map[string]any) { delete(m, "key") }),
		"null key":         mutate(func(m map[string]any) { m["key"] = nil }),
		"broken key":       mutate(func(m map[string]any) { m["key"] = "AAAA" }),
		"not base64url":    mutate(func(m map[string]any) { m["key"] = "+/+/" }),
		"unsupported key":  mutate(func(m map[string]any) { m["key"] = Base64URL(p384PKCS8(t)).String() }),
		"no credential id": mutate(func(m map[string]any) { delete(m, "credentialId") }),
		"long id":          mutate(func(m map[string]any) { m["credentialId"] = Base64URL(make([]byte, 1024)).String() }),
		"no rp id":         mutate(func(m map[string]any) { m["rpId"] = "" }),
		"rp id not a name": mutate(func(m map[string]any) { m["rpId"] = "../../etc" }),
		"upper case rp id": mutate(func(m map[string]any) { m["rpId"] = "Example.com" }),
		"wrong algorithm":  mutate(func(m map[string]any) { m["coseAlg"] = -8 }),
		"broken date":      mutate(func(m map[string]any) { m["creationDate"] = "yesterday" }),
		"no user handle":   mutate(func(m map[string]any) { delete(m, "userHandle") }),
		"long user handle": mutate(func(m map[string]any) { m["userHandle"] = Base64URL(make([]byte, 65)).String() }),
		"wrong type":       mutate(func(m map[string]any) { m["signCount"] = "many" }),
	} {
		got, err := Decode(in)
		require.ErrorIs(t, err, ErrInvalidCredential, name)
		assert.Nil(t, got, name)
	}

	// The optional members can be left out. Padding is tolerated.
	got, err := Decode(mutate(func(m map[string]any) {
		delete(m, "coseAlg")
		delete(m, "backupEligible")
		delete(m, "backupState")
		delete(m, "creationDate")
		m["userHandle"] = "aGFuZGxl"
		m["credentialId"] = "aWQ="
	}))
	require.NoError(t, err)
	assert.True(t, got.BackupEligible)
	assert.True(t, got.BackupState)
	assert.True(t, got.Created.IsZero())
	assert.Equal(t, []byte("handle"), got.UserHandle)
	assert.Equal(t, []byte("id"), got.ID)

	// The longest ID that is allowed.
	got, err = Decode(mutate(func(m map[string]any) { m["credentialId"] = Base64URL(make([]byte, 1023)).String() }))
	require.NoError(t, err)

	att, err := got.Attest(Consent{Present: true})
	require.NoError(t, err)
	assert.Len(t, parseAuthData(t, att.AuthenticatorData).credentialID, 1023)

	// A credential that can not be used is not serialized.
	for _, broken := range []*Credential{
		nil,
		{RPID: "example.com"},
		{ID: []byte("id"), RPID: "example.com", UserHandle: []byte("handle")},
		{ID: []byte("id"), RPID: "example.com", UserHandle: []byte("handle"), Key: &ecdsa.PrivateKey{}},
		{ID: []byte("id"), RPID: "example.com", UserHandle: []byte("handle"), Key: &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: elliptic.P256()}}},
		{ID: []byte("id"), RPID: "example.com", UserHandle: []byte("handle"), Key: &rsa.PrivateKey{}},
		{ID: []byte("id"), RPID: "example.com", UserHandle: []byte("handle"), Key: ed25519.PrivateKey{1, 2, 3}},
	} {
		_, err = broken.Encode()
		require.Error(t, err)

		_, err = broken.Attest(Consent{Present: true})
		require.Error(t, err)

		hash := sha256.Sum256(nil)
		_, err = broken.Assert(hash[:], Consent{Present: true})
		require.Error(t, err)
	}
}

func TestBase64URL(t *testing.T) {
	t.Parallel()

	buf, err := json.Marshal(struct {
		A Base64URL `json:"a"`
		B Base64URL `json:"b,omitempty"`
	}{A: Base64URL{0xfb, 0xff}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"a":"-_8"}`, string(buf))

	var got struct {
		A Base64URL `json:"a"`
		B Base64URL `json:"b"`
		C Base64URL `json:"c"`
		D Base64URL `json:"d"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"a":"-_8","b":"-_8=","c":null,"d":"-w=="}`), &got))
	assert.Equal(t, Base64URL{0xfb, 0xff}, got.A)
	assert.Equal(t, Base64URL{0xfb, 0xff}, got.B)
	assert.Nil(t, got.C)
	assert.Equal(t, Base64URL{0xfb}, got.D)

	parsed, err := ParseBase64URL("-_8")
	require.NoError(t, err)
	assert.Equal(t, Base64URL{0xfb, 0xff}, parsed)
	assert.Equal(t, "-_8", parsed.String())

	_, err = ParseBase64URL(`-_8"`)
	require.Error(t, err)

	for _, in := range []string{
		`"+/8"`,     // not the URL safe alphabet
		`42`,        // not a string
		`"-_8=="`,   // too much padding
		`"-_8A="`,   // padding that completes nothing
		`"-w="`,     // not enough padding
		`"=-_8"`,    // padding in the wrong place
		`"-_\n8"`,   // line breaks
		`"-_8\r\n"`, // line breaks
		`"-_ 8"`,    // spaces
		`"-_9"`,     // trailing bits that are not zero
		`"A"`,       // not a multiple of six bits
	} {
		require.Error(t, json.Unmarshal([]byte(`{"a":`+in+`}`), &got), in)
	}
}
