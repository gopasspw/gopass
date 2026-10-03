package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/stretchr/testify/require"
)

// A native envelope exercises the same keyring/agent flow as a hardware
// plugin, without requiring hardware or interactive authentication in CI.
func TestAgeRecipientKeyringSession(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("age agent Unix sockets are not available on Windows")
	}
	ts := newAgeTester(t)
	defer func() {
		_, _ = ts.run("age agent stop")
		ts.teardown()
	}()
	ts.initAgeStore(true)

	protector, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	bootstrap := filepath.Join(ts.tempDir, "keyring-unlock.txt")
	require.NoError(t, os.WriteFile(bootstrap, []byte(protector.String()+"\n"), 0o600))
	out, err := ts.runCmd([]string{ts.Binary, "config", "age.keyring-identities", bootstrap}, nil)
	require.NoError(t, err, out)
	out, err = ts.runCmd([]string{ts.Binary, "config", "age.keyring-recipients", protector.Recipient().String()}, nil)
	require.NoError(t, err, out)
	out, err = ts.run("age identities reencrypt")
	require.NoError(t, err, out)

	// Envelope migration must not change the identities used by store entries.
	t.Setenv("GOPASS_AGE_PASSWORD", "")
	t.Setenv("GOPASS_AGE_STDIN_PASSPHRASE", "")
	for _, entry := range []string{"session/first", "session/second"} {
		out, err = ts.runCmd([]string{ts.Binary, "insert", "-m", entry}, []byte("session secret for "+entry+"\n"))
		require.NoError(t, err, out)
	}
	out, err = ts.run("age agent unlock")
	require.NoError(t, err, out)
	for _, entry := range []string{"session/first", "session/second"} {
		out, err = ts.runCmd([]string{ts.Binary, "show", "-o", entry}, nil)
		require.NoError(t, err, out)
		require.Contains(t, out, "session secret for "+entry)
	}
	client := agent.NewClient()
	status, err := client.Status()
	require.NoError(t, err)
	require.Equal(t, "OK", status)
	require.NoError(t, client.Lock())
	status, err = client.Status()
	require.NoError(t, err)
	require.Equal(t, "locked", status)
	out, err = ts.run("age agent unlock")
	require.NoError(t, err, out)
	out, err = ts.runCmd([]string{ts.Binary, "show", "-o", "session/first"}, nil)
	require.NoError(t, err, out)
	require.Contains(t, out, "session secret for session/first")

	// Replacement of bootstrap credentials must invalidate a loaded session;
	// it must not keep decrypting with a software identity cached beforehand.
	wrong, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(bootstrap, []byte(wrong.String()+"\n"), 0o600))
	out, err = ts.runCmd([]string{ts.Binary, "show", "-o", "session/second"}, nil)
	require.Error(t, err, out)
	status, err = client.Status()
	require.NoError(t, err)
	require.Equal(t, "locked", status)
}
