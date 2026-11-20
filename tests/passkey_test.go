package tests

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// webauthnResponse is the part of the JSON printed by the passkey commands
// that a relying party needs to verify a ceremony.
type webauthnResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Response struct {
		ClientDataJSON    string `json:"clientDataJSON"`
		AuthenticatorData string `json:"authenticatorData"`
		PublicKey         string `json:"publicKey"`
		Signature         string `json:"signature"`
		UserHandle        string `json:"userHandle"`
	} `json:"response"`
}

func b64(t *testing.T, s string) []byte {
	t.Helper()

	buf, err := base64.RawURLEncoding.DecodeString(s)
	require.NoError(t, err)

	return buf
}

func TestPasskey(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("passkey")
	require.Error(t, err)

	ts.initStore()

	t.Run("no passkeys", func(t *testing.T) {
		out, err := ts.run("passkey")
		require.NoError(t, err)
		assert.Contains(t, out, "No passkeys found")

		_, err = ts.run("passkey list example.com")
		require.Error(t, err)
	})

	t.Run("usage", func(t *testing.T) {
		out, err := ts.run("passkey create")
		require.Error(t, err)
		assert.Equal(t, "Usage: "+filepath.Base(ts.Binary)+" passkey create <rp-id> <user>\n", out)

		out, err = ts.run("passkey assert")
		require.Error(t, err)
		assert.Contains(t, out, "Usage: "+filepath.Base(ts.Binary)+" passkey assert")
	})

	t.Run("not confirmed", func(t *testing.T) {
		_, err := ts.run("passkey create example.com alice")
		require.Error(t, err)

		_, err = ts.run("show passkeys/example.com/alice")
		require.Error(t, err)
	})

	var reg webauthnResponse

	t.Run("create", func(t *testing.T) {
		out, err := ts.run("--yes passkey create example.com alice --display-name Alice --challenge cmVnaXN0ZXI")
		require.NoError(t, err, out)
		require.NoError(t, json.Unmarshal([]byte(out), &reg), out)

		assert.Equal(t, "public-key", reg.Type)
		assert.JSONEq(t,
			`{"type":"webauthn.create","challenge":"cmVnaXN0ZXI","origin":"https://example.com","crossOrigin":false}`,
			string(b64(t, reg.Response.ClientDataJSON)))

		// A passkey can not be created for another site.
		out, err = ts.run("--yes passkey create example.com mallory --origin https://evil.example")
		require.Error(t, err)
		assert.Contains(t, out, "SecurityError")
	})

	t.Run("list", func(t *testing.T) {
		out, err := ts.run("passkey list")
		require.NoError(t, err)
		assert.Contains(t, out, "passkeys/example.com/alice")
		assert.Contains(t, out, "alice (Alice)")
		assert.Contains(t, out, reg.ID)
		assert.NotContains(t, out, "mallory")

		out, err = ts.run("ls")
		require.NoError(t, err)
		assert.Contains(t, out, "passkeys")
		assert.Contains(t, out, "example.com")
	})

	t.Run("private key is hidden", func(t *testing.T) {
		_, err := ts.run("config show.safecontent true")
		require.NoError(t, err)

		out, err := ts.run("show passkeys/example.com/alice")
		require.NoError(t, err)
		assert.Contains(t, out, "passkey: *****")
		assert.NotContains(t, out, "credentialId")

		_, err = ts.run("config show.safecontent false")
		require.NoError(t, err)

		out, err = ts.run("show passkeys/example.com/alice")
		require.NoError(t, err)
		assert.Contains(t, out, `passkey: {"v":1,"rpId":"example.com","userName":"alice","userDisplayName":"Alice",`)
	})

	t.Run("assert", func(t *testing.T) {
		pub, err := x509.ParsePKIXPublicKey(b64(t, reg.Response.PublicKey))
		require.NoError(t, err)

		key, ok := pub.(*ecdsa.PublicKey)
		require.True(t, ok)

		for _, arg := range []string{"example.com", "passkeys/example.com/alice"} {
			out, err := ts.run("--yes passkey assert " + arg + " --challenge c2lnbiBpbg --origin https://login.example.com")
			require.NoError(t, err, out)

			var auth webauthnResponse
			require.NoError(t, json.Unmarshal([]byte(out), &auth), out)
			assert.Equal(t, reg.ID, auth.ID)
			assert.NotEmpty(t, auth.Response.UserHandle)

			clientData := b64(t, auth.Response.ClientDataJSON)
			assert.JSONEq(t,
				`{"type":"webauthn.get","challenge":"c2lnbiBpbg","origin":"https://login.example.com","crossOrigin":false}`,
				string(clientData))

			// what a relying party does to verify the assertion
			authData := b64(t, auth.Response.AuthenticatorData)
			rpIDHash := sha256.Sum256([]byte("example.com"))
			require.Len(t, authData, 37)
			assert.Equal(t, rpIDHash[:], authData[:32])
			assert.Equal(t, byte(0x19), authData[32], "user present, backup eligible and backed up")
			assert.Equal(t, []byte{0, 0, 0, 0}, authData[33:], "no signature counter")

			hash := sha256.Sum256(clientData)
			digest := sha256.Sum256(append(authData, hash[:]...))
			assert.True(t, ecdsa.VerifyASN1(key, digest[:], b64(t, auth.Response.Signature)))
		}

		out, err := ts.run("--yes passkey assert example.com --challenge c2lnbiBpbg --origin https://evil.example")
		require.Error(t, err)
		assert.Contains(t, out, "SecurityError")

		_, err = ts.run("passkey assert example.com --challenge c2lnbiBpbg")
		require.Error(t, err, "not confirmed")

		_, err = ts.run("--yes passkey assert example.org --challenge c2lnbiBpbg")
		require.Error(t, err, "no passkey")
	})

	t.Run("user verification", func(t *testing.T) {
		// A confirmation is not a verification of the user.
		out, err := ts.run("--yes passkey assert example.com --challenge c2lnbiBpbg --user-verification required")
		require.Error(t, err)
		assert.Contains(t, out, "passkey.user-verification")

		// unless the user declares it to be one
		_, err = ts.run("config passkey.user-verification true")
		require.NoError(t, err)

		out, err = ts.run("--yes passkey assert example.com --challenge c2lnbiBpbg --user-verification required")
		require.NoError(t, err, out)

		var auth webauthnResponse
		require.NoError(t, json.Unmarshal([]byte(out), &auth), out)
		assert.Equal(t, byte(0x1d), b64(t, auth.Response.AuthenticatorData)[32], "user present and verified")
	})

	t.Run("export is not logged", func(t *testing.T) {
		log := filepath.Join(ts.tempDir, "debug.log")
		t.Setenv("GOPASS_DEBUG_LOG", log)

		// to standard output, with the debug log enabled
		out, err := ts.run("passkey export example.com")
		require.NoError(t, err, out)

		key := regexp.MustCompile(`"key": "([A-Za-z0-9_-]{40,})"`).FindStringSubmatch(out)
		require.Len(t, key, 2, out)

		logged, err := os.ReadFile(log)
		require.NoError(t, err)
		require.NotEmpty(t, logged)
		assert.NotContains(t, string(logged), key[1], "the private key must not end up in the debug log")
		assert.NotContains(t, string(logged), key[1][:24])

		require.NoError(t, os.Unsetenv("GOPASS_DEBUG_LOG"))
	})

	t.Run("export and import", func(t *testing.T) {
		file := filepath.Join(ts.workDir(), "passkeys.json")

		out, err := ts.run("passkey export --output " + file)
		require.NoError(t, err, out)
		assert.Contains(t, out, "not encrypted")

		_, err = ts.run("rm -f passkeys/example.com/alice")
		require.NoError(t, err)

		_, err = ts.run("passkey list example.com")
		require.Error(t, err)

		out, err = ts.run("passkey import " + file)
		require.NoError(t, err, out)
		assert.Contains(t, out, "passkeys/example.com/alice")

		out, err = ts.run("passkey list example.com")
		require.NoError(t, err)
		assert.Contains(t, out, reg.ID)
	})

	t.Run("secret named like the site", func(t *testing.T) {
		// Many stores have one. It does not hide the passkeys of the site.
		_, err := ts.runWithInput("insert -f example.com", "hunter2\n")
		require.NoError(t, err)

		out, err := ts.run("--yes passkey assert example.com --challenge c2lnbiBpbg")
		require.NoError(t, err, out)

		var auth webauthnResponse
		require.NoError(t, json.Unmarshal([]byte(out), &auth), out)
		assert.Equal(t, reg.ID, auth.ID)
	})

	t.Run("store config is ignored", func(t *testing.T) {
		// The config of a store is versioned with it. Whoever can push to
		// the store must not be able to redirect new passkeys.
		cfg := filepath.Join(ts.storeDir("root"), "config")
		require.NoError(t, os.WriteFile(cfg, []byte("[passkey]\n\tprefix = shared/passkeys\n"), 0o600))

		out, err := ts.run("--yes passkey create example.com bob")
		require.NoError(t, err, out)
		require.NoError(t, os.Remove(cfg))

		out, err = ts.run("passkey list example.com")
		require.NoError(t, err)
		assert.Contains(t, out, "passkeys/example.com/bob")
		assert.NotContains(t, out, "shared/")
	})

	t.Run("remove", func(t *testing.T) {
		// Two passkeys match, and there is nobody to pick one.
		_, err := ts.run("--yes passkey remove example.com")
		require.Error(t, err)

		out, err := ts.run("--yes passkey remove passkeys/example.com/bob")
		require.NoError(t, err, out)
		assert.Contains(t, out, "Removed the passkey")
		assert.Contains(t, out, "holds no passkeys anymore")

		out, err = ts.run("passkey list example.com")
		require.NoError(t, err)
		assert.NotContains(t, out, "passkeys/example.com/bob")
		assert.Contains(t, out, reg.ID)

		// Not confirmed: kept.
		_, err = ts.run("passkey remove example.com")
		require.Error(t, err)

		out, err = ts.run("passkey remove example.com --force --credential " + reg.ID)
		require.NoError(t, err, out)

		_, err = ts.run("passkey list example.com")
		require.Error(t, err)
	})
}

// TestPasskeyIsCommitted checks that a passkey is not just written but also
// committed if the store is under version control. Otherwise it would never
// reach another device.
func TestPasskeyIsCommitted(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("init --crypto=gpgcli --storage=gitfs " + keyID)
	require.NoError(t, err, "failed to init git store:\n%s", out)

	// with a signature counter, so that signing in writes to the store too
	_, err = ts.run("config passkey.signcount true")
	require.NoError(t, err)

	git := func(args ...string) string {
		t.Helper()

		raw, err := exec.Command("git", append([]string{"-C", ts.storeDir("root")}, args...)...).CombinedOutput()
		require.NoError(t, err, string(raw))

		return strings.TrimSpace(string(raw))
	}

	out, err = ts.run("--yes passkey create example.com alice")
	require.NoError(t, err, out)
	require.True(t, json.Valid([]byte(out)), "nothing but the response is printed:\n%s", out)

	assert.Empty(t, git("status", "--porcelain"), "the passkey is committed")
	assert.Contains(t, git("log", "-1", "--format=%s"), "passkeys/example.com/alice")
	commits := git("rev-list", "--count", "HEAD")

	out, err = ts.run("--yes passkey assert example.com --challenge c2lnbiBpbg")
	require.NoError(t, err, out)
	require.True(t, json.Valid([]byte(out)), "nothing but the response is printed:\n%s", out)

	assert.Empty(t, git("status", "--porcelain"), "the counter is committed")
	assert.NotEqual(t, commits, git("rev-list", "--count", "HEAD"))
}
