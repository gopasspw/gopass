package age

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"filippo.io/age/plugin"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHardwareKeyringUnlockSession(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("age agent requires Unix sockets")
	}
	useShortTempDir(t)
	a := newTestAge(t)
	dir := t.TempDir()
	executable := filepath.Join(dir, "age-plugin-gopasstest")
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", executable, "./agent/testdata/plugin")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	countFile := filepath.Join(dir, "count")
	t.Setenv("GOPASS_TEST_PLUGIN_COUNT", countFile)
	count := func() int {
		data, err := os.ReadFile(countFile)
		if os.IsNotExist(err) {
			return 0
		}
		require.NoError(t, err)

		return strings.Count(string(data), "unwrap\n")
	}
	software := keyringTestIdentity(t)
	protector := keyringTestIdentity(t)
	bootstrap := filepath.Join(dir, "unlock.txt")
	writePlugin := func(mode string) {
		line := plugin.EncodeIdentity("gopasstest", []byte(mode+"|"+protector.String()))
		require.NoError(t, os.WriteFile(bootstrap, []byte(line+"\n"), 0o600))
	}
	writePlugin("decrypt")
	ctx := protectedKeyringContext(t, bootstrap, protector.Recipient().String())
	require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))
	require.Equal(t, 1, count(), "new envelope must be verified through bootstrap before replacing the keyring")
	cfg, _ := config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("age.agent-enabled", "true"))
	require.NoError(t, cfg.SetEnv("age.agent-timeout", "1"))
	startFreshAgent(t)
	client := agent.NewClient()
	t.Cleanup(func() { require.NoError(t, client.Quit()) })
	first := encryptToRecipient(t, software.Recipient(), []byte("first entry"))
	second := encryptToRecipient(t, software.Recipient(), []byte("second entry"))
	plaintext, err := a.Decrypt(ctx, first)
	require.NoError(t, err)
	assert.Equal(t, "first entry", string(plaintext))
	plaintext, err = a.Decrypt(ctx, second)
	require.NoError(t, err)
	assert.Equal(t, "second entry", string(plaintext))
	require.Equal(t, 2, count(), "multiple entries must unwrap the protected keyring only once")
	source, err := a.sessionFingerprint(ctx)
	require.NoError(t, err)
	plaintext, err = client.DecryptSession(second, source)
	require.NoError(t, err)
	assert.Equal(t, "second entry", string(plaintext))
	require.NoError(t, client.Lock())
	_, err = client.DecryptSession(first, source)
	require.Error(t, err, "raw agent must not fall back to bootstrap after locking")
	_, err = a.Decrypt(ctx, first)
	require.NoError(t, err)
	require.Equal(t, 3, count(), "lock must require another bootstrap authentication")
	require.Eventually(t, func() bool {
		status, e := client.Status()

		return e == nil && status == "locked"
	}, 3*time.Second, 20*time.Millisecond)
	_, err = client.DecryptSession(second, source)
	require.Error(t, err)
	_, err = a.Decrypt(ctx, second)
	require.NoError(t, err)
	require.Equal(t, 4, count(), "timeout must require another bootstrap authentication")

	// A changed bootstrap must discard the old session before a failed unlock.
	writePlugin("cancel")
	plaintext, err = a.Decrypt(ctx, second)
	require.ErrorContains(t, err, "test authentication canceled")
	require.Nil(t, plaintext)
	require.Equal(t, 5, count(), "authentication failure must not trigger a second fallback attempt")
	status, err := client.Status()
	require.NoError(t, err)
	assert.Equal(t, "locked", status)
	_, err = client.DecryptSession(first, source)
	require.Error(t, err, "old software credentials must not survive canceled authentication")
	writePlugin("decrypt")
	_, err = a.Decrypt(ctx, first)
	require.NoError(t, err)
	require.Equal(t, 6, count(), "a failed unlock must leave the agent usable")
	// Missing bootstrap is also an invalid source and must clear active keys.
	require.NoError(t, os.Remove(bootstrap))
	_, err = a.Decrypt(ctx, first)
	require.Error(t, err)
	status, err = client.Status()
	require.NoError(t, err)
	assert.Equal(t, "locked", status)
}

func setupHardwareKeyring(t *testing.T) (*Age, context.Context, string) {
	t.Helper()
	useShortTempDir(t)
	a := newTestAge(t)
	a.identity = filepath.Join(appdir.UserConfig(), "age", "identities")
	dir := t.TempDir()
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", filepath.Join(dir, "age-plugin-gopasstest"), "./agent/testdata/plugin")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", out)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	protector := keyringTestIdentity(t)
	software := keyringTestIdentity(t)
	bootstrap := filepath.Join(dir, "unlock.txt")
	encoding := plugin.EncodeIdentity("gopasstest", []byte("decrypt|"+protector.String()))
	require.NoError(t, os.WriteFile(bootstrap, []byte(encoding+"\n"), 0o600))
	ctx := protectedKeyringContext(t, bootstrap, protector.Recipient().String())
	require.NoError(t, a.saveIdentities(ctx, []string{software.String()}, true))

	return a, ctx, software.Recipient().String()
}

func TestHardwareKeyringAutostartDoesNotAuthenticate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("age agent requires Unix sockets")
	}
	_, ctx, recipient := setupHardwareKeyring(t)
	countFile := filepath.Join(t.TempDir(), "count")
	t.Setenv("GOPASS_TEST_PLUGIN_COUNT", countFile)
	cfg, _ := config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("age.agent-enabled", "true"))
	launches := 0
	ctx = WithAgentLauncher(ctx, func(context.Context) error {
		launches++
		startFreshAgent(t)

		return nil
	})
	a, err := New(ctx, false, "")
	require.NoError(t, err)
	client := agent.NewClient()
	t.Cleanup(func() { require.NoError(t, client.Quit()) })
	require.Equal(t, 1, launches)
	_, err = os.Stat(countFile)
	require.True(t, os.IsNotExist(err), "constructing backend must not request hardware authentication")
	recs, err := a.parseRecipients(ctx, []string{recipient})
	require.NoError(t, err)
	ciphertext := encryptToRecipient(t, recs[0], []byte("autostart entry"))
	plaintext, err := a.Decrypt(ctx, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, "autostart entry", string(plaintext))
	data, err := os.ReadFile(countFile)
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(string(data), "unwrap\n"), "first read must authenticate once, without eager startup authentication")
}

func TestHardwareKeyringLockDuringAuthentication(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("age agent requires Unix sockets")
	}
	a, ctx, recipient := setupHardwareKeyring(t)
	cfg, _ := config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("age.agent-enabled", "true"))
	startFreshAgent(t)
	client := agent.NewClient()
	t.Cleanup(func() { require.NoError(t, client.Quit()) })
	recs, err := a.parseRecipients(ctx, []string{recipient})
	require.NoError(t, err)
	ciphertext := encryptToRecipient(t, recs[0], []byte("hardware entry"))
	dir := t.TempDir()
	ready, release := filepath.Join(dir, "ready"), filepath.Join(dir, "release")
	t.Setenv("GOPASS_TEST_PLUGIN_READY", ready)
	t.Setenv("GOPASS_TEST_PLUGIN_RELEASE", release)
	result := make(chan error, 1)
	go func() { _, err := a.Decrypt(ctx, ciphertext); result <- err }()
	require.Eventually(t, func() bool {
		_, err := os.Stat(ready)

		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	// This models a screen lock while a hardware prompt is still pending.
	require.NoError(t, client.Lock())
	require.NoError(t, os.WriteFile(release, []byte("released"), 0o600))
	select {
	case err := <-result:
		require.Error(t, err, "completion after lock must not reload identities")
	case <-time.After(5 * time.Second):
		t.Fatal("hardware operation did not finish")
	}
	status, err := client.Status()
	require.NoError(t, err)
	require.Equal(t, "locked", status)
	source, err := a.sessionFingerprint(ctx)
	require.NoError(t, err)
	_, err = client.DecryptSession(ciphertext, source)
	require.Error(t, err, "agent must remain locked after stale authentication finishes")
}
