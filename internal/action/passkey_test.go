package action

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/passkey"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// passkeyTest sets up an action with an empty store and captured output.
func passkeyTest(t *testing.T, yes bool) (context.Context, *Action, *bytes.Buffer) {
	t.Helper()

	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, yes)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithTerminal(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = &bytes.Buffer{}
	stdin = strings.NewReader("")

	t.Cleanup(func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdin = os.Stdin
	})

	return ctx, act, buf
}

func TestPasskeyCreateAndAssert(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	// nothing there yet
	require.NoError(t, act.PasskeyList(ctx, gptest.CliCtx(ctx, t)))
	assert.Contains(t, buf.String(), "No passkeys found")
	buf.Reset()

	err := act.PasskeyList(ctx, gptest.CliCtx(ctx, t, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))
	buf.Reset()

	// usage errors
	for _, args := range [][]string{{}, {"example.com"}} {
		err := act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, args...))
		require.Error(t, err)
		assert.Equal(t, exit.Usage, exitCode(t, err))
	}

	err = act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES512"}, "example.com", "alice"))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	err = act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256", "challenge": "+/+/"}, "example.com", "alice"))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	// A passkey is bound to the site it is created for.
	err = act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256", "origin": "https://evil.example"}, "example.com", "alice"))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))
	assert.False(t, act.Store.Exists(ctx, "passkeys/example.com/alice"))
	buf.Reset()

	// create
	flags := map[string]string{
		"algorithm":    "EdDSA",
		"display-name": "Alice A",
		"challenge":    "cmVnaXN0ZXI",
		"origin":       "https://login.example.com",
	}
	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "example.com", "alice")))

	var reg passkey.RegistrationResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &reg), buf.String())
	assert.Equal(t, passkey.EdDSA, reg.Response.PublicKeyAlgorithm)
	assert.JSONEq(t,
		`{"type":"webauthn.create","challenge":"cmVnaXN0ZXI","origin":"https://login.example.com","crossOrigin":false}`,
		string(reg.Response.ClientDataJSON))
	buf.Reset()

	// stored as a single line below passkeys/<rp id>/<user>
	sec, err := act.Store.Get(ctx, "passkeys/example.com/alice")
	require.NoError(t, err)

	creds, err := passkey.FromSecret(sec)
	require.NoError(t, err)
	require.Len(t, creds, 1)
	assert.Equal(t, "example.com", creds[0].RPID)
	assert.Equal(t, "alice", creds[0].UserName)
	assert.Equal(t, "Alice A", creds[0].UserDisplayName)
	assert.Equal(t, []byte(reg.RawID), creds[0].ID)
	assert.Zero(t, creds[0].SignCount)

	// list
	require.NoError(t, act.PasskeyList(ctx, gptest.CliCtx(ctx, t)))
	assert.Contains(t, buf.String(), "passkeys/example.com/alice")
	assert.Contains(t, buf.String(), "alice (Alice A)")
	assert.Contains(t, buf.String(), "EdDSA")
	assert.Contains(t, buf.String(), reg.ID)
	buf.Reset()

	require.NoError(t, act.PasskeyList(ctx, gptest.CliCtx(ctx, t, "example.com")))
	assert.Contains(t, buf.String(), reg.ID)
	buf.Reset()

	// assert, selecting the passkey by the RP ID or by the secret
	pub, err := x509.ParsePKIXPublicKey(reg.Response.PublicKey)
	require.NoError(t, err)

	for _, name := range []string{"example.com", "passkeys/example.com/alice"} {
		flags := map[string]string{"challenge": "c2lnbiBpbg", "user-verification": "preferred"}
		require.NoError(t, act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, name)), name)

		var auth passkey.AuthenticationResponse
		require.NoError(t, json.Unmarshal(buf.Bytes(), &auth), buf.String())
		assert.Equal(t, reg.ID, auth.ID)
		assert.JSONEq(t,
			`{"type":"webauthn.get","challenge":"c2lnbiBpbg","origin":"https://example.com","crossOrigin":false}`,
			string(auth.Response.ClientDataJSON))

		hash := sha256.Sum256(auth.Response.ClientDataJSON)
		signed := append(append([]byte{}, auth.Response.AuthenticatorData...), hash[:]...)

		edPub, ok := pub.(ed25519.PublicKey)
		require.True(t, ok)
		assert.True(t, ed25519.Verify(edPub, signed, auth.Response.Signature))
		buf.Reset()
	}

	// usage errors
	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"challenge": "c2lnbiBpbg"}))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	err = act.PasskeyAssert(ctx, gptest.CliCtx(ctx, t, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"challenge": "c2lnbiBpbg", "origin": "https://evil.example"}, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"challenge": "c2lnbiBpbg"}, "example.org"))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))

	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"challenge": "c2lnbiBpbg", "credential": "dW5rbm93bg"}, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))
}

func TestPasskeyCreateTwice(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	create := func(user string, flags map[string]string) passkey.RegistrationResponse {
		t.Helper()
		buf.Reset()

		all := map[string]string{"algorithm": "ES256"}
		for k, v := range flags {
			all[k] = v
		}

		require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, all, "example.com", user)))

		var reg passkey.RegistrationResponse
		require.NoError(t, json.Unmarshal(buf.Bytes(), &reg), buf.String())
		buf.Reset()

		return reg
	}

	first := create("alice", nil)
	second := create("alice", nil)
	assert.NotEqual(t, first.ID, second.ID)

	// Creating a passkey for a user again adds one for the same account.
	// The first one is kept: it may be the only one the site accepts.
	sec, err := act.Store.Get(ctx, "passkeys/example.com/alice")
	require.NoError(t, err)

	creds, err := passkey.FromSecret(sec)
	require.NoError(t, err)
	require.Len(t, creds, 2)
	assert.Equal(t, []byte(first.RawID), creds[0].ID)
	assert.Equal(t, []byte(second.RawID), creds[1].ID)
	assert.Equal(t, creds[0].UserHandle, creds[1].UserHandle)

	// A passkey for another account that happens to have the same name.
	third := create("alice", map[string]string{"user-handle": "YW5vdGhlcg"})

	sec, err = act.Store.Get(ctx, "passkeys/example.com/alice")
	require.NoError(t, err)

	creds, err = passkey.FromSecret(sec)
	require.NoError(t, err)
	require.Len(t, creds, 3)
	assert.Equal(t, []byte("another"), creds[2].UserHandle)

	// Now the secret alone does not identify a passkey anymore.
	flags := map[string]string{"challenge": "c2lnbiBpbg"}
	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "passkeys/example.com/alice"))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	for _, name := range []string{"example.com", "passkeys/example.com/alice"} {
		buf.Reset()

		flags := map[string]string{"challenge": "c2lnbiBpbg", "credential": third.ID}
		require.NoError(t, act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, name)))

		var auth passkey.AuthenticationResponse
		require.NoError(t, json.Unmarshal(buf.Bytes(), &auth), buf.String())
		assert.Equal(t, third.ID, auth.ID)
		assert.Equal(t, passkey.Base64URL("another"), auth.Response.UserHandle)
	}
}

func TestPasskeyFlags(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, "example.com", "alice")))
	buf.Reset()

	// A flag that is given without a value is a mistake, not a request for
	// the default.
	for _, flag := range []string{"challenge", "user-handle"} {
		flags := map[string]string{"algorithm": "ES256", flag: ""}
		err := act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "example.com", "bob"))
		require.Error(t, err, flag)
		assert.Equal(t, exit.Usage, exitCode(t, err), flag)
	}

	assert.False(t, act.Store.Exists(ctx, "passkeys/example.com/bob"))

	for name, flags := range map[string]map[string]string{
		"no challenge":          {"origin": "https://example.com"},
		"empty challenge":       {"challenge": ""},
		"empty credential":      {"challenge": "c2lnbiBpbg", "credential": ""},
		"malformed credential":  {"challenge": "c2lnbiBpbg", "credential": "+/+/"},
		"unknown verification":  {"challenge": "c2lnbiBpbg", "user-verification": "always"},
		"malformed challenge":   {"challenge": "not base64url!"},
		"origin of other site":  {"challenge": "c2lnbiBpbg", "origin": "https://evil.example"},
		"origin is not a tuple": {"challenge": "c2lnbiBpbg", "origin": "example.com"},
	} {
		err := act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "example.com"))
		require.Error(t, err, name)
		assert.Equal(t, exit.Usage, exitCode(t, err), name)
		assert.Empty(t, buf.String(), name)
	}

	err := act.PasskeyExport(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"output": ""}))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))
	assert.Empty(t, buf.String(), "the private keys are not printed by accident")

	// RP IDs are not case sensitive on the command line.
	flags := map[string]string{"challenge": "c2lnbiBpbg"}
	require.NoError(t, act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "Example.COM")))
	buf.Reset()

	require.NoError(t, act.PasskeyList(ctx, gptest.CliCtx(ctx, t, "EXAMPLE.com")))
	assert.Contains(t, buf.String(), "passkeys/example.com/alice")
}

func TestPasskeyListIsSafeToPrint(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	// The names of an account are chosen by a web site.
	name := "eve\x1b[2J\nSECRET\tRP ID\u202e"
	flags := map[string]string{"algorithm": "ES256", "display-name": "Eve\x07\r\nof\tEvil"}
	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "example.com", name)))
	buf.Reset()

	require.NoError(t, act.PasskeyList(ctx, gptest.CliCtx(ctx, t)))

	// one header, one passkey, no control characters
	assert.Equal(t, 2, strings.Count(buf.String(), "\n"), buf.String())
	assert.Contains(t, buf.String(), "eve [2J SECRET RP ID (Eve of Evil)")

	for _, r := range buf.String() {
		assert.True(t, r == '\n' || unicode.IsPrint(r), "%q", r)
	}

	// The names are stored as the site gave them.
	entries, err := act.passkeyH.store(ctx).List(ctx, "example.com")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, name, entries[0].Credential.UserName)
	assert.Equal(t, "passkeys/example.com/eve[2JSECRETRP ID", entries[0].Name)
}

func TestPasskeyNeedsConfirmation(t *testing.T) {
	ctx, act, buf := passkeyTest(t, false)

	// Nobody is there to confirm: nothing is created.
	err := act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, "example.com", "alice"))
	require.Error(t, err)
	assert.Equal(t, exit.Aborted, exitCode(t, err))
	assert.False(t, act.Store.Exists(ctx, "passkeys/example.com/alice"))
	assert.Empty(t, buf.String())

	yes := ctxutil.WithAlwaysYes(ctx, true)
	require.NoError(t, act.PasskeyCreate(yes, gptest.CliCtxWithFlags(yes, t, map[string]string{"algorithm": "ES256"}, "example.com", "alice")))
	buf.Reset()

	// and nothing is signed
	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"challenge": "c2lnbiBpbg"}, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.Aborted, exitCode(t, err))
	assert.Empty(t, buf.String())
}

func TestPasskeyUserVerificationConfig(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	require.NoError(t, act.cfg.Set("", "passkey.signcount", "true"))

	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, "example.com", "alice")))
	buf.Reset()

	// A confirmation is no verification. A site that insists on one can not
	// be served.
	flags := map[string]string{"challenge": "c2lnbiBpbg", "user-verification": "required"}
	err := act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "example.com"))
	require.Error(t, err)
	assert.Empty(t, buf.String())

	var last uint32

	for _, verified := range []bool{false, true} {
		buf.Reset()

		// Unless the user declares using the store to be a verification.
		if verified {
			require.NoError(t, act.cfg.Set("", "passkey.user-verification", "true"))
		} else {
			flags["user-verification"] = "preferred"
		}

		require.NoError(t, act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, flags, "example.com")))
		flags["user-verification"] = "required"

		var auth passkey.AuthenticationResponse
		require.NoError(t, json.Unmarshal(buf.Bytes(), &auth), buf.String())

		ad := auth.Response.AuthenticatorData
		require.Len(t, ad, 37)
		assert.True(t, passkey.Flags(ad[32]).Has(passkey.FlagUserPresent))
		assert.Equal(t, verified, passkey.Flags(ad[32]).Has(passkey.FlagUserVerified))

		// The signature counter was opted in to and is persisted.
		count := uint32(ad[33])<<24 | uint32(ad[34])<<16 | uint32(ad[35])<<8 | uint32(ad[36])
		assert.Greater(t, count, last)
		last = count

		sec, err := act.Store.Get(ctx, "passkeys/example.com/alice")
		require.NoError(t, err)

		creds, err := passkey.FromSecret(sec)
		require.NoError(t, err)
		require.Len(t, creds, 1)
		assert.Equal(t, count, creds[0].SignCount)
	}
}

func TestPasskeyExportAndImport(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	err := act.PasskeyExport(ctx, gptest.CliCtx(ctx, t))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))

	for _, user := range []string{"alice", "bob"} {
		require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, "example.com", user)))
	}

	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "RS256"}, "example.org", "carol")))
	buf.Reset()

	// to standard output, limited to one relying party
	require.NoError(t, act.PasskeyExport(ctx, gptest.CliCtx(ctx, t, "example.org")))
	assert.Contains(t, buf.String(), `"rpId": "example.org"`)
	assert.NotContains(t, buf.String(), `"rpId": "example.com"`)
	assert.True(t, json.Valid(buf.Bytes()), buf.String())
	buf.Reset()

	// to a file only the user can read
	file := filepath.Join(t.TempDir(), "passkeys.json")
	require.NoError(t, act.PasskeyExport(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"output": file})))

	fi, err := os.Stat(file)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		// Windows reports writable files as 0o666; the folder's ACL
		// determines who can read the export.
		assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
	}

	exported, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, 3, strings.Count(string(exported), `"type": "passkey"`))

	// An existing file is not replaced.
	err = act.PasskeyExport(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"output": file}))
	require.Error(t, err)
	assert.Equal(t, exit.IO, exitCode(t, err))

	// import into an empty store
	before, err := act.passkeyH.store(ctx).List(ctx, "")
	require.NoError(t, err)
	require.Len(t, before, 3)

	for _, e := range before {
		require.NoError(t, act.Store.Delete(ctx, e.Name))
	}

	err = act.PasskeyImport(ctx, gptest.CliCtx(ctx, t))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	err = act.PasskeyImport(ctx, gptest.CliCtx(ctx, t, filepath.Join(t.TempDir(), "missing.json")))
	require.Error(t, err)
	assert.Equal(t, exit.IO, exitCode(t, err))

	buf.Reset()
	require.NoError(t, act.PasskeyImport(ctx, gptest.CliCtx(ctx, t, file)))
	assert.Contains(t, buf.String(), "Imported 3 passkeys")

	// importing again, this time from standard input, changes nothing
	stdin = bytes.NewReader(exported)
	require.NoError(t, act.PasskeyImport(ctx, gptest.CliCtx(ctx, t, "-")))

	after, err := act.passkeyH.store(ctx).List(ctx, "")
	require.NoError(t, err)
	require.Len(t, after, 3)

	for i := range before {
		assert.Equal(t, before[i].Name, after[i].Name)
		assert.Equal(t, before[i].Credential.ID, after[i].Credential.ID)
		assert.Equal(t, before[i].Credential.UserHandle, after[i].Credential.UserHandle)
		assert.Equal(t, before[i].Credential.Algorithm(), after[i].Credential.Algorithm())
	}

	// garbage
	stdin = strings.NewReader("not a credential exchange document")
	err = act.PasskeyImport(ctx, gptest.CliCtx(ctx, t, "-"))
	require.Error(t, err)

	// The format has no signature counter. A passkey that maintains one is
	// left out, and an export without any passkey is not written.
	require.NoError(t, act.cfg.Set("", "passkey.signcount", "true"))
	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, "counting.example", "dave")))
	buf.Reset()

	counting := filepath.Join(t.TempDir(), "counting.json")
	err = act.PasskeyExport(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"output": counting}, "Counting.Example"))
	require.Error(t, err)
	assert.Equal(t, exit.Unsupported, exitCode(t, err))
	assert.NoFileExists(t, counting)

	require.NoError(t, act.PasskeyExport(ctx, gptest.CliCtx(ctx, t)))
	assert.Equal(t, 3, strings.Count(buf.String(), `"type": "passkey"`))
	assert.NotContains(t, buf.String(), "counting.example")
}

func TestPasskeyIsHiddenByShow(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, "example.com", "alice")))
	buf.Reset()

	sec, err := act.Store.Get(ctx, "passkeys/example.com/alice")
	require.NoError(t, err)

	// The private key is not shown if safecontent is enabled.
	assert.True(t, isUnsafeKey(ctx, passkey.SecretKey, sec))
	assert.Equal(t, "passkey: *****", showSafeContent(ctx, sec))
}

// passkeyCreate creates a passkey and returns the registration response.
func passkeyCreate(ctx context.Context, t *testing.T, act *Action, buf *bytes.Buffer, rpID, user string) passkey.RegistrationResponse {
	t.Helper()

	buf.Reset()
	require.NoError(t, act.PasskeyCreate(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"algorithm": "ES256"}, rpID, user)))

	var reg passkey.RegistrationResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &reg), buf.String())
	buf.Reset()

	return reg
}

// passkeyIDs returns the IDs of the passkeys in a secret.
func passkeyIDs(ctx context.Context, t *testing.T, act *Action, name string) []string {
	t.Helper()

	sec, err := act.Store.Get(ctx, name)
	require.NoError(t, err)

	creds, err := passkey.FromSecret(sec)
	require.NoError(t, err)

	ids := make([]string, 0, len(creds))
	for _, c := range creds {
		ids = append(ids, passkey.Base64URL(c.ID).String())
	}

	return ids
}

func TestPasskeyAssertSecretOrRPID(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	reg := passkeyCreate(ctx, t, act, buf, "example.com", "alice")

	assertWith := func(name string, flags map[string]string) passkey.AuthenticationResponse {
		t.Helper()
		buf.Reset()

		all := map[string]string{"challenge": "c2lnbiBpbg"}
		for k, v := range flags {
			all[k] = v
		}

		require.NoError(t, act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, all, name)), name)

		var auth passkey.AuthenticationResponse
		require.NoError(t, json.Unmarshal(buf.Bytes(), &auth), buf.String())
		buf.Reset()

		return auth
	}

	// Many stores have a secret named after a site. It does not hide the
	// passkeys of the site.
	login := secrets.New()
	login.SetPassword("hunter2")
	require.NoError(t, act.Store.Set(ctx, "example.com", login))

	assert.Equal(t, reg.ID, assertWith("example.com", nil).ID)

	// A passkey can be kept anywhere, e.g. next to the password, and is
	// found by the name of its secret.
	stored, err := act.Store.Get(ctx, "passkeys/example.com/alice")
	require.NoError(t, err)

	line, _ := stored.Get(passkey.SecretKey)
	require.NoError(t, login.Set(passkey.SecretKey, line))
	require.NoError(t, act.Store.Set(ctx, "websites/example.com", login))

	auth := assertWith("websites/example.com", map[string]string{"origin": "https://login.example.com"})
	assert.Equal(t, reg.ID, auth.ID)
	assert.Contains(t, string(auth.Response.ClientDataJSON), `"origin":"https://login.example.com"`)

	// The secret named like the site holds a passkey now. It is used.
	require.NoError(t, act.Store.Set(ctx, "example.com", login))
	assert.Equal(t, reg.ID, assertWith("example.com", nil).ID)

	// A secret that holds no passkey is no RP ID if it can not be one.
	plain := secrets.New()
	plain.SetPassword("hunter2")
	require.NoError(t, act.Store.Set(ctx, "websites/example.org", plain))

	err = act.PasskeyAssert(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"challenge": "c2lnbiBpbg"}, "websites/example.org"))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))
	assert.Empty(t, buf.String())
}

func TestPasskeyRemove(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	first := passkeyCreate(ctx, t, act, buf, "example.com", "alice")
	second := passkeyCreate(ctx, t, act, buf, "example.com", "alice")
	passkeyCreate(ctx, t, act, buf, "example.com", "bob")

	alice := "passkeys/example.com/alice"
	require.Equal(t, []string{first.ID, second.ID}, passkeyIDs(ctx, t, act, alice))

	err := act.PasskeyRemove(ctx, gptest.CliCtx(ctx, t))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	// Which one? There is nobody to ask.
	err = act.PasskeyRemove(ctx, gptest.CliCtx(ctx, t, alice))
	require.Error(t, err)
	assert.Equal(t, exit.Usage, exitCode(t, err))

	err = act.PasskeyRemove(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"credential": "dW5rbm93bg"}, alice))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))

	// Nobody confirmed it: nothing is removed.
	no := ctxutil.WithAlwaysYes(ctx, false)
	err = act.PasskeyRemove(no, gptest.CliCtxWithFlags(no, t, map[string]string{"credential": first.ID}, alice))
	require.Error(t, err)
	assert.Equal(t, exit.Aborted, exitCode(t, err))
	assert.Equal(t, []string{first.ID, second.ID}, passkeyIDs(ctx, t, act, alice))

	// by the secret
	buf.Reset()
	require.NoError(t, act.PasskeyRemove(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"credential": first.ID}, alice)))
	assert.Equal(t, []string{second.ID}, passkeyIDs(ctx, t, act, alice))
	assert.Contains(t, buf.String(), "Removed the passkey")
	assert.NotContains(t, buf.String(), "holds no passkeys")

	// by the RP ID, without asking
	buf.Reset()
	require.NoError(t, act.PasskeyRemove(no, gptest.CliCtxWithFlags(no, t, map[string]string{"credential": second.ID, "force": "true"}, "example.com")))
	assert.Empty(t, passkeyIDs(ctx, t, act, alice))
	assert.Contains(t, buf.String(), alice+" holds no passkeys anymore")

	// The secret is kept. It may hold more than passkeys.
	assert.True(t, act.Store.Exists(ctx, alice))

	// The only passkey left needs no selection.
	require.NoError(t, act.PasskeyRemove(ctx, gptest.CliCtx(ctx, t, "example.com")))

	err = act.PasskeyList(ctx, gptest.CliCtx(ctx, t, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))

	err = act.PasskeyRemove(ctx, gptest.CliCtx(ctx, t, "example.com"))
	require.Error(t, err)
	assert.Equal(t, exit.NotFound, exitCode(t, err))
}

func TestPasskeyExportCountsEachPasskeyOnce(t *testing.T) {
	ctx, act, buf := passkeyTest(t, true)

	passkeyCreate(ctx, t, act, buf, "example.com", "alice")

	// a copy in a shared folder
	sec, err := act.Store.Get(ctx, "passkeys/example.com/alice")
	require.NoError(t, err)
	require.NoError(t, act.Store.Set(ctx, "team/passkeys/example.com/alice", sec))

	file := filepath.Join(t.TempDir(), "passkeys.json")
	require.NoError(t, act.PasskeyExport(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"output": file})))
	assert.Contains(t, buf.String(), "Exported 1 passkeys")

	exported, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(exported), `"type": "passkey"`))
}

func TestPasskeyChoice(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 11, 20, 18, 51, 15, 0, time.UTC)
	entry := func(id string) passkey.Entry {
		return passkey.Entry{
			Name: "passkeys/example.com/alice",
			Credential: &passkey.Credential{
				ID:              []byte(id),
				UserName:        "alice",
				UserDisplayName: "Alice",
				Created:         created,
			},
		}
	}

	// Two passkeys of one account, in one secret, can be told apart.
	assert.Equal(t, "alice (Alice) - passkeys/example.com/alice (created 2026-11-20, ID MDEyMzQ1…)", passkeyChoice(entry("0123456789")))
	assert.NotEqual(t, passkeyChoice(entry("0123456789")), passkeyChoice(entry("9876543210")))

	short := entry("ab")
	short.Credential.Created = time.Time{}
	assert.Equal(t, "alice (Alice) - passkeys/example.com/alice (ID YWI)", passkeyChoice(short))
}

// exitCode returns the exit code of an error returned by an action.
func exitCode(t *testing.T, err error) int {
	t.Helper()

	var ee cli.ExitCoder
	require.ErrorAs(t, err, &ee)

	return ee.ExitCode()
}
