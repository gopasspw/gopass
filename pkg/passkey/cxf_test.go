package passkey

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCXFRoundTrip(t *testing.T) {
	t.Parallel()

	entries := make([]Entry, 0, 4)

	for i, alg := range SupportedAlgorithms() {
		cred, err := New(NewOptions{
			RPID:            fmt.Sprintf("site%d.example", i),
			UserHandle:      []byte{0xfb, 0xff, byte(i)},
			UserName:        "alice",
			UserDisplayName: "Alice <A&B>",
			Algorithm:       alg,
		})
		require.NoError(t, err)

		entries = append(entries, Entry{Name: "passkeys/" + cred.RPID + "/alice", Credential: cred})
	}

	entries[0].Credential.Extensions = json.RawMessage(`{"payments":true}`)

	// The format has no signature counter: such passkeys are not exported.
	counting, err := New(NewOptions{RPID: "counting.example", UserHandle: []byte("handle"), SignCounter: true})
	require.NoError(t, err)
	entries = append(entries, Entry{Name: "passkeys/counting.example/alice", Credential: counting})

	var buf bytes.Buffer

	exported, skipped, err := ExportCXF(&buf, entries)
	require.NoError(t, err)
	require.Len(t, skipped, 1)
	assert.Equal(t, counting.ID, skipped[0].Credential.ID)
	assert.Len(t, exported, len(entries)-1)

	// the structure of the document
	var doc struct {
		Version struct {
			Major *int `json:"major"`
			Minor *int `json:"minor"`
		} `json:"version"`
		ExporterRPID        string `json:"exporterRpId"`
		ExporterDisplayName string `json:"exporterDisplayName"`
		Timestamp           int64  `json:"timestamp"`
		Accounts            []struct {
			ID          string            `json:"id"`
			Username    *string           `json:"username"`
			Email       *string           `json:"email"`
			Collections []json.RawMessage `json:"collections"`
			Items       []struct {
				ID          string           `json:"id"`
				CreationAt  int64            `json:"creationAt"`
				Title       string           `json:"title"`
				Subtitle    string           `json:"subtitle"`
				Credentials []map[string]any `json:"credentials"`
			} `json:"items"`
		} `json:"accounts"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &doc))

	require.NotNil(t, doc.Version.Major)
	require.NotNil(t, doc.Version.Minor)
	assert.Equal(t, 1, *doc.Version.Major)
	assert.Equal(t, 0, *doc.Version.Minor)
	assert.Equal(t, "gopass.pw", doc.ExporterRPID)
	assert.Equal(t, "gopass", doc.ExporterDisplayName)
	assert.InDelta(t, time.Now().Unix(), doc.Timestamp, 60)
	require.Len(t, doc.Accounts, 1)

	account := doc.Accounts[0]
	assert.NotEmpty(t, account.ID)
	// required members are present, required arrays even if they are empty
	require.NotNil(t, account.Username)
	require.NotNil(t, account.Email)
	require.NotNil(t, account.Collections)
	assert.Contains(t, buf.String(), `"collections": []`)
	require.Len(t, account.Items, 3)

	for i, item := range account.Items {
		cred := entries[i].Credential

		assert.NotEmpty(t, item.ID)
		assert.Equal(t, cred.RPID, item.Title)
		assert.Equal(t, "alice (Alice <A&B>)", item.Subtitle)
		assert.Equal(t, cred.Created.Unix(), item.CreationAt)
		require.Len(t, item.Credentials, 1)

		pk := item.Credentials[0]
		assert.Equal(t, "passkey", pk["type"])
		assert.Equal(t, Base64URL(cred.ID).String(), pk["credentialId"])
		assert.Equal(t, cred.RPID, pk["rpId"])
		assert.Equal(t, "alice", pk["username"])
		assert.Equal(t, "Alice <A&B>", pk["userDisplayName"])
		assert.Equal(t, Base64URL(cred.UserHandle).String(), pk["userHandle"])
		assert.NotEmpty(t, pk["key"])
		assert.NotContains(t, pk, "signCount")
	}

	assert.Equal(t, map[string]any{"payments": true}, account.Items[0].Credentials[0]["fido2Extensions"])
	assert.NotContains(t, account.Items[1].Credentials[0], "fido2Extensions")

	// and back
	creds, err := ImportCXF(&buf)
	require.NoError(t, err)
	require.Len(t, creds, 3)

	for i, got := range creds {
		want := entries[i].Credential

		assert.Equal(t, want.ID, got.ID)
		assert.Equal(t, want.RPID, got.RPID)
		assert.Equal(t, want.UserHandle, got.UserHandle)
		assert.Equal(t, want.UserName, got.UserName)
		assert.Equal(t, want.UserDisplayName, got.UserDisplayName)
		assert.Equal(t, want.Algorithm(), got.Algorithm())
		assert.True(t, want.Created.Equal(got.Created))
		assert.Zero(t, got.SignCount)
		assert.True(t, got.BackupEligible)
		assert.True(t, got.BackupState)

		wantKey, err := want.PublicKey()
		require.NoError(t, err)
		gotKey, err := got.PublicKey()
		require.NoError(t, err)
		assert.Equal(t, wantKey, gotKey)
	}

	assert.JSONEq(t, `{"payments":true}`, string(creds[0].Extensions))
	assert.Empty(t, creds[1].Extensions)
}

func TestImportCXF(t *testing.T) {
	t.Parallel()

	good, err := New(NewOptions{RPID: "webauthn.example", UserHandle: []byte("handle"), UserName: "johndoe"})
	require.NoError(t, err)

	key, err := marshalPrivateKey(good.Key)
	require.NoError(t, err)

	p384, err := json.Marshal(Base64URL(p384PKCS8(t)))
	require.NoError(t, err)

	// A document as other providers write it: with members and credential
	// types gopass does not know, and with padded base64url.
	doc := `{
  "version": {"major": 1, "minor": 3},
  "exporterRpId": "exporter.example.com",
  "exporterDisplayName": "Exporter app",
  "timestamp": 1705228800,
  "futureMember": [1, 2, 3],
  "accounts": [
    {
      "id": "DZSXp7iBQY-Fg-OofakQtQ",
      "username": "jane_smith",
      "email": "jane.smith@example.com",
      "fullName": "Jane Smith",
      "collections": [{"id": "c29tZQ", "title": "Work", "items": []}],
      "items": [
        {
          "id": "9OF-QjVDQo2Wp2xWPw6ZhA",
          "title": "Login",
          "scope": {"urls": ["https://webauthn.example"], "androidApps": []},
          "credentials": [
            {"type": "basic-auth", "username": {"fieldType": "string", "value": "johndoe"}},
            {"type": "totp", "secret": "JBSWY3DPEHPK3PXP", "period": 30, "digits": 6, "algorithm": "sha256"},
            {"type": "note", "content": {"fieldType": "string", "value": "passkey"}},
            {"type": "from-the-future", "key": 42, "rpId": ["not", "a", "string"]},
            "not even an object"
          ],
          "tags": ["work"]
        },
        {
          "id": "akKA3Y0jQRuK7sKplB0Y9w",
          "creationAt": 1705142400,
          "modifiedAt": 1705228800,
          "title": "WebAuthn",
          "subtitle": "johndoe",
          "favorite": true,
          "credentials": [
            {
              "type": "passkey",
              "credentialId": "Y3JlZGVudGlhbElkRXhhbXBsZQ==",
              "rpId": "webauthn.example",
              "username": "johndoe",
              "userDisplayName": "John Doe",
              "userHandle": "cnEzaNHWcYK3coWZjvoaV1Hj9gnI12mKe2dL2HZVFlY",
              "key": "` + Base64URL(key).String() + `",
              "fido2Extensions": {
                "hmacCredentials": {"algorithm": "hmac-sha256", "credWithUV": "AAAA", "credWithoutUV": "AAAA"},
                "payments": true
              }
            },
            {
              "type": "passkey",
              "credentialId": "dW5zdXBwb3J0ZWQ",
              "rpId": "p384.example",
              "username": "johndoe",
              "userDisplayName": "",
              "userHandle": "aGFuZGxl",
              "key": ` + string(p384) + `
            },
            {
              "type": "passkey",
              "credentialId": "bm9rZXk",
              "rpId": "nokey.example",
              "username": "johndoe",
              "userDisplayName": "",
              "userHandle": "aGFuZGxl",
              "key": ""
            },
            {
              "type": "passkey",
              "credentialId": "YmFkcnA",
              "rpId": "../../etc",
              "username": "johndoe",
              "userDisplayName": "",
              "userHandle": "aGFuZGxl",
              "key": "` + Base64URL(key).String() + `"
            }
          ]
        }
      ]
    }
  ]
}`

	creds, err := ImportCXF(strings.NewReader(doc))

	// Passkeys that can not be used are reported, not silently dropped.
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedKey)
	require.ErrorIs(t, err, ErrInvalidCredential)
	assert.Contains(t, err.Error(), "p384.example")
	assert.Contains(t, err.Error(), "nokey.example")
	assert.Contains(t, err.Error(), "../../etc")

	require.Len(t, creds, 1)

	got := creds[0]
	assert.Equal(t, []byte("credentialIdExample"), got.ID)
	assert.Equal(t, "webauthn.example", got.RPID)
	assert.Equal(t, "johndoe", got.UserName)
	assert.Equal(t, "John Doe", got.UserDisplayName)
	assert.Len(t, got.UserHandle, 32)
	assert.Equal(t, ES256, got.Algorithm())
	assert.Equal(t, int64(1705142400), got.Created.Unix())
	assert.Zero(t, got.SignCount)
	assert.False(t, got.Counting())
	assert.Contains(t, string(got.Extensions), "hmacCredentials")

	want, err := good.PublicKey()
	require.NoError(t, err)
	pub, err := got.PublicKey()
	require.NoError(t, err)
	assert.Equal(t, want, pub)

	// The extension data survives being stored.
	line, err := got.Encode()
	require.NoError(t, err)

	stored, err := Decode(line)
	require.NoError(t, err)
	assert.JSONEq(t, string(got.Extensions), string(stored.Extensions))
}

func TestImportCXFInvalid(t *testing.T) {
	t.Parallel()

	_, err := ImportCXF(strings.NewReader("not json"))
	require.Error(t, err)

	_, err = ImportCXF(strings.NewReader(`{"version":{"major":2,"minor":0},"accounts":[]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2.0")

	creds, err := ImportCXF(strings.NewReader(`{"version":{"major":1,"minor":0},"accounts":[]}`))
	require.NoError(t, err)
	assert.Empty(t, creds)

	_, _, err = ExportCXF(&bytes.Buffer{}, []Entry{{Name: "broken", Credential: &Credential{RPID: "example.com"}}})
	require.ErrorIs(t, err, ErrInvalidCredential)

	// more than one document
	_, err = ImportCXF(strings.NewReader(`{"version":{"major":1,"minor":0},"accounts":[]} {"accounts":[]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected data")
}

func TestImportCXFMalformedPasskeys(t *testing.T) {
	t.Parallel()

	good, err := New(NewOptions{RPID: "example.com", UserHandle: []byte("handle"), UserName: "alice"})
	require.NoError(t, err)

	key := Base64URL(must(marshalPrivateKey(good.Key))).String()

	passkey := func(mutate func(m map[string]any)) string {
		m := map[string]any{
			"type":            "passkey",
			"credentialId":    "aWQ",
			"rpId":            "example.com",
			"username":        "alice",
			"userDisplayName": "Alice",
			"userHandle":      "aGFuZGxl",
			"key":             key,
		}
		mutate(m)

		return string(must(json.Marshal(m)))
	}

	// A passkey that is malformed is reported. It is not mistaken for a
	// credential of a type that is unknown, which would be skipped silently.
	for name, pk := range map[string]string{
		"bad base64":       passkey(func(m map[string]any) { m["credentialId"] = "+/+/" }),
		"wrong type":       passkey(func(m map[string]any) { m["rpId"] = []string{"example.com"} }),
		"null key":         passkey(func(m map[string]any) { m["key"] = nil }),
		"no key":           passkey(func(m map[string]any) { delete(m, "key") }),
		"no credential id": passkey(func(m map[string]any) { delete(m, "credentialId") }),
		"no rp id":         passkey(func(m map[string]any) { delete(m, "rpId") }),
		"no user handle":   passkey(func(m map[string]any) { delete(m, "userHandle") }),
		"null user handle": passkey(func(m map[string]any) { m["userHandle"] = nil }),
		// The RP ID is part of what a site verifies. It is not rewritten.
		"upper case rp id": passkey(func(m map[string]any) { m["rpId"] = "Example.COM" }),
		"unicode rp id":    passkey(func(m map[string]any) { m["rpId"] = "bücher.example" }),
		"rp id is an ip":   passkey(func(m map[string]any) { m["rpId"] = "192.0.2.1" }),
	} {
		doc := `{"version":{"major":1,"minor":0},"accounts":[{"items":[{"title":"Item","credentials":[` +
			pk + `,` + passkey(func(map[string]any) {}) + `]}]}]}`

		creds, err := ImportCXF(strings.NewReader(doc))
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), `skipping passkey of item "Item"`, name)

		// the passkey next to it is imported
		require.Len(t, creds, 1, name)
		assert.Equal(t, []byte("id"), creds[0].ID, name)
	}
}

func TestExportCXFIdentifiers(t *testing.T) {
	t.Parallel()

	cred := newCred(t, "example.com", "alice", "handle")
	other := newCred(t, "example.com", "bob", "other")

	// The same passkey can be stored twice, e.g. as a copy in a shared
	// store. It is exported, and counted, once.
	entries := []Entry{
		{Name: "passkeys/example.com/alice", Credential: cred},
		{Name: "team/passkeys/example.com/alice", Credential: cred},
		{Name: "passkeys/example.com/bob", Credential: other},
	}

	export := func() (string, []string) {
		var buf bytes.Buffer

		exported, skipped, err := ExportCXF(&buf, entries)
		require.NoError(t, err)
		assert.Empty(t, skipped)
		require.Len(t, exported, 2)
		assert.Equal(t, "passkeys/example.com/alice", exported[0].Name)
		assert.Equal(t, "passkeys/example.com/bob", exported[1].Name)

		var doc struct {
			Accounts []struct {
				ID    string `json:"id"`
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
			} `json:"accounts"`
		}
		require.NoError(t, json.Unmarshal(buf.Bytes(), &doc))
		require.Len(t, doc.Accounts, 1)
		require.Len(t, doc.Accounts[0].Items, 2)

		items := []string{doc.Accounts[0].Items[0].ID, doc.Accounts[0].Items[1].ID}
		assert.NotEqual(t, items[0], items[1])

		return doc.Accounts[0].ID, items
	}

	// Identifiers are the same in every export, so that an importer can
	// recognize what it has seen before.
	account, items := export()
	account2, items2 := export()
	assert.Equal(t, account, account2)
	assert.Equal(t, items, items2)

	id, err := ParseBase64URL(account)
	require.NoError(t, err)
	assert.Len(t, id, 16)
}
