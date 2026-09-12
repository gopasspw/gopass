//go:build linux

package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// installPath returns a path inside the tester's temporary home directory, which
// GOPASS_HOMEDIR points at.
func installPath(ts *tester, elem ...string) string {
	return filepath.Join(append([]string{ts.tempDir}, elem...)...)
}

// TestSecretServiceInstallUninstall verifies the install subcommand writes the
// systemd user unit and the D-Bus session activation file, and that uninstall
// removes them again. GOPASS_HOMEDIR redirects the XDG paths into the test's
// temporary directory.
func TestSecretServiceInstallUninstall(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	systemdUnit := installPath(ts, ".config", "systemd", "user", "gopass-secret-service.service")
	activation := installPath(ts, ".local", "share", "dbus-1", "services", "org.freedesktop.secrets.service")

	out, err := ts.run("secret-service install")
	require.NoError(t, err, "secret-service install failed:\n%s", out)
	assert.Contains(t, out, systemdUnit)

	unit, err := os.ReadFile(systemdUnit)
	require.NoError(t, err, "systemd unit not written")
	assert.Contains(t, string(unit), "ExecStart=")
	assert.Contains(t, string(unit), "secret-service serve")
	assert.Contains(t, string(unit), "BusName=org.freedesktop.secrets")

	body, err := os.ReadFile(activation)
	require.NoError(t, err, "D-Bus activation file not written")
	assert.Contains(t, string(body), "Name=org.freedesktop.secrets")

	out, err = ts.run("secret-service uninstall")
	require.NoError(t, err, "secret-service uninstall failed:\n%s", out)

	_, err = os.Stat(systemdUnit)
	assert.True(t, os.IsNotExist(err), "systemd unit still present")
	_, err = os.Stat(activation)
	assert.True(t, os.IsNotExist(err), "D-Bus activation file still present")

	// A second uninstall must be a no-op.
	out, err = ts.run("secret-service uninstall")
	require.NoError(t, err, "second uninstall failed:\n%s", out)
	assert.Contains(t, out, "Nothing to remove")
}

// TestSecretServiceSecretTool verifies interoperability with a real
// libsecret-based client. secret-tool negotiates the
// "dh-ietf1024-sha256-aes128-cbc-pkcs7" transport, so a successful round-trip
// proves the DH implementation is wire-compatible with libsecret (not just
// with our own tests).
func TestSecretServiceSecretTool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if _, err := exec.LookPath("secret-tool"); err != nil {
		t.Skipf("secret-tool (libsecret) not available: %v", err)
	}

	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	address, stopBus := startPrivateBus(t)
	defer stopBus()

	t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)

	logs := startSecretServiceDaemon(t, ts)

	conn := connectBus(t, address)
	waitForSecretService(t, conn, logs)

	// runTool executes secret-tool with a timeout so a hang cannot stall CI.
	runTool := func(stdin string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "secret-tool", args...)
		if stdin != "" {
			cmd.Stdin = strings.NewReader(stdin)
		}

		out, err := cmd.CombinedOutput()

		return strings.TrimSpace(string(out)), err
	}

	attrs := []string{"service", "gopass-integration-test"}

	// Store a secret through libsecret.
	out, err := runTool("hunter2", append([]string{"store", "--label", "Gopass Test"}, attrs...)...)
	require.NoError(t, err, "secret-tool store failed: %s", out)

	// The secret must now be visible in the gopass store.
	ls, err := ts.run("ls secret-service")
	require.NoError(t, err, "gopass ls failed: %s", ls)
	assert.Contains(t, ls, "default", "collection not created in the gopass store:\n%s", ls)

	// Look it up through libsecret. Lookup returns the value on stdout, so the
	// combined output is compared exactly.
	cmd := exec.Command("secret-tool", "lookup", attrs[0], attrs[1])
	cmd.Env = append(cmd.Environ(), "DBUS_SESSION_BUS_ADDRESS="+address)

	looked, err := cmd.Output()
	require.NoError(t, err, "secret-tool lookup failed")
	assert.Equal(t, "hunter2", strings.TrimSpace(string(looked)))

	// Clear it again.
	out, err = runTool("", append([]string{"clear"}, attrs...)...)
	require.NoError(t, err, "secret-tool clear failed: %s", out)

	// Lookup must now fail with a non-zero exit code.
	if out, err := runTool("", "lookup", attrs[0], attrs[1]); err == nil {
		t.Fatalf("secret-tool lookup succeeded after clear: %s", out)
	}
}
