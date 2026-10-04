//go:build !windows

package tests

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/stretchr/testify/require"
)

// A native envelope exercises the same keyring/agent flow as a hardware
// plugin, with a pseudo-terminal for manual recipient approval in CI.
func TestAgeRecipientKeyringSession(t *testing.T) {
	// GOPASS_HOMEDIR determines the socket directory even when XDG_RUNTIME_DIR
	// is set. Keep the test home below macOS's Unix socket path limit.
	probe := filepath.Join(os.TempDir(), t.Name()+strings.Repeat("x", 10), "001", ".run", "gopass-age-agent.sock")
	if len(probe) > 100 {
		t.Setenv("TMPDIR", "/tmp")
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
	original, err := os.ReadFile(filepath.Join(ts.tempDir, ".config", "gopass", "age", "identities"))
	require.NoError(t, err)
	out, err = ts.run("age identities reencrypt")
	require.Error(t, err, "noninteractive migration must not silently approve recipients: %s", out)
	require.Contains(t, out, "manual review")
	out, err = runKeyringReview(t, ts, "n\n")
	require.Error(t, err, out)
	unchanged, err := os.ReadFile(filepath.Join(ts.tempDir, ".config", "gopass", "age", "identities"))
	require.NoError(t, err)
	require.Equal(t, original, unchanged)
	out, err = runKeyringReview(t, ts, "y\n")
	require.NoError(t, err, out)
	require.Contains(t, out, protector.Recipient().String())

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

// runKeyringReview supplies terminal input rather than bypassing production
// confirmation. The existing pty dependency provides Unix PTYs for this test.
func runKeyringReview(t *testing.T, ts *tester, answer string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ts.Binary, "age", "identities", "reencrypt")
	cmd.Dir = ts.workDir()
	terminal, err := pty.Start(cmd)
	require.NoError(t, err)
	defer func() { _ = terminal.Close() }()
	_, err = terminal.WriteString(answer)
	require.NoError(t, err)
	output, readErr := io.ReadAll(terminal)
	// Linux reports EIO when the slave closes; other Unix systems return EOF.
	if readErr != nil && !errors.Is(readErr, syscall.EIO) {
		require.NoError(t, readErr)
	}

	return string(output), cmd.Wait()
}
