package age

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

// useShortTempDir relocates TMPDIR to a short, writable directory before any
// t.TempDir() call, but only if the default TMPDIR would produce a socket path
// too long for the OS. macOS caps unix-domain socket paths at ~104 bytes, and a
// deeply nested TMPDIR (e.g. under nix-shell) plus the appended
// ".run/gopass-age-agent.sock" easily exceeds that, surfacing as
// `bind: invalid argument` when the agent tries to listen.
func useShortTempDir(t *testing.T) {
	t.Helper()
	// Worst-case probe: a long temp dir name (t.TempDir appends the test name
	// plus a random suffix) followed by the socket suffix.
	probe := filepath.Join(os.TempDir(), strings.Repeat("x", 48), ".run", "gopass-age-agent.sock")
	if len(probe) <= 100 {
		return
	}
	for _, c := range []string{"/tmp/claude/gpa", "/tmp/gpa"} {
		if err := os.MkdirAll(c, 0o700); err == nil {
			t.Setenv("TMPDIR", c)

			return
		}
	}
	t.Fatal("cannot find a short writable TMPDIR for the age agent unix socket")
}

// ctxWithAgentEnabled returns a context whose config has age.agent-enabled=true
// so that Age.Decrypt routes through the agent code path.
func ctxWithAgentEnabled(t *testing.T) context.Context {
	t.Helper()
	cfg := config.NewInMemory()
	require.NoError(t, cfg.Set("", "age.agent-enabled", "true"))

	return cfg.WithConfig(t.Context())
}

// startFreshAgent starts an age agent that — like one launched ahead of time by
// launchd — holds NO identities, and waits until it is reachable. The socket is
// isolated to the test's GOPASS_HOMEDIR (set up by newTestAge).
func startFreshAgent(t *testing.T) *agent.Agent {
	t.Helper()
	ctx := t.Context()
	a, err := agent.New()
	require.NoError(t, err)

	runErr := make(chan error, 1)
	go func() {
		runErr <- a.Run(ctx)
	}()

	c := agent.NewClient()
	var lastPingErr error
	for range 60 {
		if err := c.Ping(); err == nil {
			return a
		} else {
			lastPingErr = err
		}
		select {
		case e := <-runErr:
			if e != nil {
				t.Fatalf("agent.Run exited early: %v", e)
			}
		default:
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("age agent did not become ready in time; last ping error: %v", lastPingErr)

	return nil
}

func encryptToRecipient(t *testing.T, rec age.Recipient, plaintext []byte) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	wc, err := age.Encrypt(buf, rec)
	require.NoError(t, err)
	_, err = wc.Write(plaintext)
	require.NoError(t, err)
	require.NoError(t, wc.Close())

	return buf.Bytes()
}

// TestDecryptSelfHealsOnFreshAgent reproduces the launchd-style failure.
//
// When the age agent is started ahead of time (e.g. by launchd) it is reachable
// but holds zero identities. Age.Decrypt is expected to notice this and load the
// backend's identities into the agent before decrypting through it.
//
// On current master this test FAILS: decryptWithAgent only self-heals when the
// agent reports "agent is locked", but a fresh *unlocked* agent returns
// "no identities specified". That error is not recognised, so Age.Decrypt
// silently falls back to direct decryption and the agent is never populated.
// The final raw client decrypt below therefore fails — which is the bug.
//
// After fixing decryptWithAgent to treat "no identities specified" the same as
// "agent is locked", the agent gets populated and the whole test passes.
func TestDecryptSelfHealsOnFreshAgent(t *testing.T) {
	useShortTempDir(t)
	a := newTestAge(t)
	ctx := ctxWithAgentEnabled(t)

	// Give the backend an identity it can load from disk (scrypt-protected with
	// the fixed test passphrase configured by newTestAge).
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	require.NoError(t, a.addIdentity(ctx, id))

	// Start a launchd-style agent: up and reachable, but with zero identities.
	ag := startFreshAgent(t)
	defer ag.Shutdown(ctx)

	plaintext := []byte("the owls are not what they seem")
	ciphertext := encryptToRecipient(t, id.Recipient(), plaintext)

	// Age.Decrypt must succeed ...
	pt, err := a.Decrypt(ctx, ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, pt)

	// ... AND it must have routed through the agent. Proof: a raw client decrypt
	// must now succeed, because Age.Decrypt should have loaded the identities
	// into the previously-empty agent. If this fails with "no identities
	// specified", Age.Decrypt fell back to direct decryption instead of using
	// the agent — the very symptom seen with a launchd-managed agent.
	raw := agent.NewClient()
	pt2, err := raw.Decrypt(ciphertext)
	require.NoError(t, err, "agent should hold identities after Age.Decrypt self-heals")
	require.Equal(t, plaintext, pt2)
}
