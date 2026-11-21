package passkey

import (
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretCodec(t *testing.T) {
	t.Parallel()

	sec := secrets.New()
	sec.SetPassword("hunter2")
	require.NoError(t, sec.Set("url", "https://example.com"))

	creds, err := FromSecret(sec)
	require.NoError(t, err)
	assert.Empty(t, creds)

	a := newCred(t, "example.com", "alice", "handle")
	b := newCred(t, "example.com", "alice", "handle")
	require.NoError(t, ToSecret(sec, []*Credential{a, b}))

	creds, err = FromSecret(sec)
	require.NoError(t, err)
	require.Len(t, creds, 2)
	assert.Equal(t, a.ID, creds[0].ID)
	assert.Equal(t, b.ID, creds[1].ID)

	// replacing the credentials leaves everything else alone
	require.NoError(t, ToSecret(sec, []*Credential{b}))

	creds, err = FromSecret(sec)
	require.NoError(t, err)
	require.Len(t, creds, 1)
	assert.Equal(t, b.ID, creds[0].ID)
	assert.Equal(t, "hunter2", sec.Password())

	url, _ := sec.Get("url")
	assert.Equal(t, "https://example.com", url)

	require.NoError(t, ToSecret(sec, nil))
	assert.Equal(t, "hunter2\nurl: https://example.com\n", string(sec.Bytes()))

	require.NoError(t, sec.Add(SecretKey, "not json"))
	_, err = FromSecret(sec)
	require.ErrorIs(t, err, ErrInvalidCredential)

	// a credential that can not be serialized does not clobber the secret
	require.Error(t, ToSecret(sec, []*Credential{{RPID: "example.com"}}))
	_, found := sec.Get(SecretKey)
	assert.True(t, found)
}

func TestSecretCodecFormats(t *testing.T) {
	t.Parallel()

	yaml, err := secrets.ParseYAML([]byte("hunter2\n---\nusername: alice\n"))
	require.NoError(t, err)
	creds, err := FromSecret(yaml)
	require.NoError(t, err)
	assert.Empty(t, creds)

	cred, err := New(NewOptions{
		RPID:            "example.com",
		UserHandle:      []byte("handle"),
		UserName:        "user: name ",
		UserDisplayName: " The: User\twith  spaces ",
		Algorithm:       RS256,
	})
	require.NoError(t, err)

	// A secret that was assembled, not parsed, may lack the final line
	// break. The passkey must not end up at the end of its last line.
	raw := secrets.NewAKV()
	_, err = raw.Write([]byte("hunter2\nurl: https://example.com"))
	require.NoError(t, err)
	require.NoError(t, ToSecret(raw, []*Credential{cred}))

	// The line survives being written and read, whatever the names contain.
	parsed := secrets.ParseAKV(raw.Bytes())
	assert.Equal(t, "hunter2", parsed.Password())

	url, _ := parsed.Get("url")
	assert.Equal(t, "https://example.com", url)

	creds, err = FromSecret(parsed)
	require.NoError(t, err)
	require.Len(t, creds, 1)
	assert.Equal(t, cred.ID, creds[0].ID)
	assert.Equal(t, "user: name ", creds[0].UserName)
	assert.Equal(t, " The: User\twith  spaces ", creds[0].UserDisplayName)
	assert.Equal(t, RS256, creds[0].Algorithm())
}

func TestAccountName(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"alice":                 "alice",
		"alice@example.com":     "alice@example.com",
		"Alice Müller":          "Alice Müller",
		"たなか":                   "たなか",
		"Алиса":                 "Алиса",
		"a/b":                   "a_b",
		`a\b`:                   "a_b",
		"../../etc/passwd":      "_.._etc_passwd",
		"..":                    "",
		".hidden":               "hidden",
		" padded ":              "padded",
		"trailing.":             "trailing",
		"line\nbreak\ttab":      "linebreaktab",
		"nul\x00byte":           "nulbyte",
		`a<b>c:d"e|f?g*h`:       "a_b_c_d_e_f_g_h",
		"":                      "",
		"   ":                   "",
		"...":                   "",
		"\u202e" + "evil":       "evil",
		strings.Repeat("ä", 80): strings.Repeat("ä", 50),
		// nothing is left dangling where a long name is cut off
		strings.Repeat("a", 99) + " b": strings.Repeat("a", 99),
		strings.Repeat("a", 99) + ".b": strings.Repeat("a", 99),
		// names of devices are reserved on Windows, with any extension
		"con":         "_con",
		"NUL":         "_NUL",
		"nul.txt":     "_nul.txt",
		"COM1":        "_COM1",
		"lpt9.backup": "_lpt9.backup",
		"console":     "console",
		"com10":       "com10",
	} {
		assert.Equal(t, want, sanitizeName(in), "%q", in)
	}

	cred := newCred(t, "example.com", "", "handle")
	assert.Equal(t, Base64URL(cred.ID).String(), accountName(cred))

	cred.UserDisplayName = "Alice"
	assert.Equal(t, "Alice", accountName(cred))

	cred.UserName = "../alice"
	assert.Equal(t, "_alice", accountName(cred))

	// Imported credentials can have IDs that are too long for a file name.
	cred.UserName, cred.UserDisplayName = "", ""
	cred.ID = make([]byte, 1023)
	assert.Len(t, accountName(cred), 43)

	// RP IDs that would escape the folder are not written
	ctx := testContext(t)
	api := apimock.New()

	for _, rpID := range []string{"", "..", "a/b", "../../outside", ".example.com", "example.com.", `a\b`, "Example.com"} {
		cred := newCred(t, "example.com", "alice", "handle")
		cred.RPID = rpID

		_, err := NewStore(ctx, api).Put(ctx, cred)
		require.ErrorIs(t, err, ErrInvalidCredential, rpID)
	}

	names, err := api.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, names)

	// A long RP ID is fine.
	long := strings.Repeat("a", 60) + "." + strings.Repeat("b", 60) + "." + strings.Repeat("c", 60) + ".example"
	name, err := NewStore(ctx, api).Put(ctx, newCred(t, long, "alice", "handle"))
	require.NoError(t, err)
	assert.Equal(t, "passkeys/"+long+"/alice", name)
}
