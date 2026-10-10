package agent

import (
	"bufio"
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/internal/buildinfo"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestAgent(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	// own the socket dir: without this every test binds the shared default
	// path, and a delayed cleanup goroutine from a previous test's quit
	// can remove the next test's freshly-listened socket
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	// start agent
	a, err := New()
	require.NoError(t, err)

	go func() {
		_ = a.Run(ctx)
	}()
	defer a.Shutdown(ctx)

	// wait for it to be ready
	time.Sleep(time.Second)

	// create client
	c := NewClient()
	require.NoError(t, c.Ping())

	// test decrypt
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	plaintext := []byte("hello world")
	buf := &bytes.Buffer{}
	wc, err := age.Encrypt(buf, id.Recipient())
	require.NoError(t, err)
	_, _ = wc.Write(plaintext)
	require.NoError(t, wc.Close())
	ciphertext := buf.Bytes()

	require.NoError(t, c.SendIdentities(id.String()))

	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// test lock
	require.NoError(t, c.Lock())

	// test quit
	require.NoError(t, c.Quit())
}

func TestAgentAutoLock(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	// own the socket dir: without this every test binds the shared default
	// path, and a delayed cleanup goroutine from a previous test's quit
	// can remove the next test's freshly-listened socket
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	// start agent
	a, err := New()
	require.NoError(t, err)

	go func() {
		_ = a.Run(ctx)
	}()
	defer a.Shutdown(ctx)

	// wait for it to be ready
	time.Sleep(time.Second)

	// create client
	c := NewClient()
	require.NoError(t, c.Ping())

	// set timeout
	require.NoError(t, c.SetTimeout(1))

	// test decrypt
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	plaintext := []byte("hello world")
	buf := &bytes.Buffer{}
	wc, err := age.Encrypt(buf, id.Recipient())
	require.NoError(t, err)
	_, _ = wc.Write(plaintext)
	require.NoError(t, wc.Close())
	ciphertext := buf.Bytes()

	require.NoError(t, c.SendIdentities(id.String()))

	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// wait for auto-lock
	time.Sleep(2 * time.Second)

	// check if locked
	_, err = c.Decrypt(ciphertext)
	require.Error(t, err)
	require.Contains(t, err.Error(), "agent is locked")

	// test quit
	require.NoError(t, c.Quit())
}

func TestAgentUnlock(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	// own the socket dir: without this every test binds the shared default
	// path, and a delayed cleanup goroutine from a previous test's quit
	// can remove the next test's freshly-listened socket
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	// start agent
	a, err := New()
	require.NoError(t, err)

	go func() {
		_ = a.Run(ctx)
	}()
	defer a.Shutdown(ctx)

	// wait for it to be ready
	time.Sleep(time.Second)

	// create client
	c := NewClient()
	require.NoError(t, c.Ping())

	// generate identity and encrypt test message
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	plaintext := []byte("hello world")
	buf := &bytes.Buffer{}
	wc, err := age.Encrypt(buf, id.Recipient())
	require.NoError(t, err)
	_, _ = wc.Write(plaintext)
	require.NoError(t, wc.Close())
	ciphertext := buf.Bytes()

	// send identities
	require.NoError(t, c.SendIdentities(id.String()))

	// verify decrypt works
	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// lock the agent
	require.NoError(t, c.Lock())

	// verify locked agent returns error
	_, err = c.Decrypt(ciphertext)
	require.Error(t, err)
	require.Contains(t, err.Error(), "agent is locked")

	// unlock the agent
	require.NoError(t, c.Unlock())

	// send identities again (simulating what unlock CLI command does)
	require.NoError(t, c.SendIdentities(id.String()))

	// verify decrypt works after unlock
	decrypted, err = c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// cleanup
	require.NoError(t, c.Quit())
}

// TestAgentMultipleIdentities verifies the agent loads ALL identities supplied
// in a single space-separated "identities" command, not just the first one on
// the line. Regression test for the framing bug where only the first identity
// was ever parsed.
func TestAgentMultipleIdentities(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	// own the socket dir: without this every test binds the shared default
	// path, and a delayed cleanup goroutine from a previous test's quit
	// can remove the next test's freshly-listened socket
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	// start agent
	a, err := New()
	require.NoError(t, err)

	go func() {
		_ = a.Run(ctx)
	}()
	defer a.Shutdown(ctx)

	// wait for it to be ready
	time.Sleep(time.Second)

	// create client
	c := NewClient()
	require.NoError(t, c.Ping())

	id1, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	id2, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	// Encrypt to id2 only: this proves the agent loaded BOTH identities, since
	// id1 cannot unwrap this ciphertext.
	plaintext := []byte("hello world")
	buf := &bytes.Buffer{}
	wc, err := age.Encrypt(buf, id2.Recipient())
	require.NoError(t, err)
	_, _ = wc.Write(plaintext)
	require.NoError(t, wc.Close())
	ciphertext := buf.Bytes()

	// Send both identities, space-separated on a single line.
	require.NoError(t, c.SendIdentities(id1.String()+" "+id2.String()))

	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// cleanup
	require.NoError(t, c.Quit())
}

func TestAgentSSHIdentities(t *testing.T) {
	ctx := t.Context()

	// own the socket dir: without this every test binds the shared default
	// path, and a delayed cleanup goroutine from a previous test's quit
	// can remove the next test's freshly-listened socket
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	a, err := New()
	require.NoError(t, err)

	go func() {
		_ = a.Run(ctx)
	}()
	defer a.Shutdown(ctx)
	time.Sleep(time.Second)

	c := NewClient()
	require.NoError(t, c.Ping())

	// Load a native identity first. Adding SSH identities must append to, not
	// replace, the identities already held by the agent.
	native, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	require.NoError(t, c.SendIdentities(native.String()))

	edPublic, edPrivate, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	edSSHPublic, err := ssh.NewPublicKey(edPublic)
	require.NoError(t, err)
	edRecipient, err := agessh.NewEd25519Recipient(edSSHPublic)
	require.NoError(t, err)

	rsaPrivate, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	rsaSSHPublic, err := ssh.NewPublicKey(&rsaPrivate.PublicKey)
	require.NoError(t, err)
	rsaRecipient, err := agessh.NewRSARecipient(rsaSSHPublic)
	require.NoError(t, err)

	identities := []struct {
		name       string
		privateKey crypto.PrivateKey
		recipient  age.Recipient
	}{
		{name: "ed25519", privateKey: edPrivate, recipient: edRecipient},
		{name: "rsa", privateKey: rsaPrivate, recipient: rsaRecipient},
	}

	for _, tc := range identities {
		t.Run(tc.name, func(t *testing.T) {
			block, err := ssh.MarshalPrivateKey(tc.privateKey, "test")
			require.NoError(t, err)
			privateKey := pem.EncodeToMemory(block)
			require.NoError(t, c.SendSSHIdentity(privateKey))
			clear(privateKey)

			plaintext := []byte("ssh identity " + tc.name)
			ciphertext := encryptForTest(t, tc.recipient, plaintext)
			decrypted, err := c.Decrypt(ciphertext)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)
		})
	}

	// The pre-existing native identity must still be available after both SSH
	// identities were appended.
	nativePlaintext := []byte("native identity remains loaded")
	nativeCiphertext := encryptForTest(t, native.Recipient(), nativePlaintext)
	decrypted, err := c.Decrypt(nativeCiphertext)
	require.NoError(t, err)
	require.Equal(t, nativePlaintext, decrypted)

	// Locking clears native and SSH identities together.
	require.NoError(t, c.Lock())
	require.NoError(t, c.Unlock())
	_, err = c.Decrypt(nativeCiphertext)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no identities specified")
}

func encryptForTest(t *testing.T, recipient age.Recipient, plaintext []byte) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	wc, err := age.Encrypt(buf, recipient)
	require.NoError(t, err)
	_, err = wc.Write(plaintext)
	require.NoError(t, err)
	require.NoError(t, wc.Close())

	return buf.Bytes()
}

func TestAgentLargePayload(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	// own the socket dir: without this every test binds the shared default
	// path, and a delayed cleanup goroutine from a previous test's quit
	// can remove the next test's freshly-listened socket
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	// start agent
	a, err := New()
	require.NoError(t, err)

	go func() {
		_ = a.Run(ctx)
	}()
	defer a.Shutdown(ctx)

	// wait for it to be ready
	time.Sleep(time.Second)

	// create client
	c := NewClient()
	require.NoError(t, c.Ping())

	// generate identity and encrypt a plaintext large enough that the
	// base64-encoded ciphertext exceeds the default 64 KiB scanner limit
	// (see issue #3508)
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	plaintext := bytes.Repeat([]byte("a"), 128*1024)
	buf := &bytes.Buffer{}
	wc, err := age.Encrypt(buf, id.Recipient())
	require.NoError(t, err)
	_, _ = wc.Write(plaintext)
	require.NoError(t, wc.Close())
	ciphertext := buf.Bytes()

	require.NoError(t, c.SendIdentities(id.String()))

	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	require.NoError(t, c.Quit())
}

// TestAgentMisframedLine simulates the framing bug documented at
// identitiesToString in the age package: a client sending identities
// newline-separated turns every identity after the first into a bare
// AGE-SECRET-KEY-1... line, which must be rejected as an unknown command
// (and, per the redaction in handleConnection, never logged verbatim).
func TestAgentMisframedLine(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	// own the socket dir: without this every test binds the shared default
	// path, and a delayed cleanup goroutine from a previous test's quit
	// can remove the next test's freshly-listened socket
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	// start agent
	a, err := New()
	require.NoError(t, err)

	go func() {
		_ = a.Run(ctx)
	}()
	defer a.Shutdown(ctx)

	// wait for it to be ready
	time.Sleep(time.Second)

	// create client
	c := NewClient()
	require.NoError(t, c.Ping())
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	require.NoError(t, c.SendIdentities(id.String()))
	plaintext := []byte("identity survives an unsupported command")
	ciphertext := encryptForTest(t, id.Recipient(), plaintext)

	_, err = c.send("AGE-SECRET-KEY-1ZZMISFRAMEZZEXAMPLEKEY")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown command")

	// This is also the compatibility contract for a new client talking to an
	// old agent: rejection of an unknown identity command must not alter the
	// identities the agent already holds.
	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// cleanup
	require.NoError(t, c.Quit())
}

// shortTempDir returns a scratch directory whose path stays well under
// the unix socket path limit even on hosts with a long TMPDIR (macOS
// /var/folders/...): darwin caps sun_path at 104 bytes, linux at 108, and
// the socket lives another ~35 bytes below this directory. GOPASS_HOMEDIR
// isolation is used instead of XDG_RUNTIME_DIR because windows resolves
// the runtime dir from LOCALAPPDATA and would ignore the XDG variable.
// Pinning GOPASS_HOMEDIR per test also gives every test its own socket
// path, so tests cannot delete each other's sockets through the agent's
// quit cleanup.
func shortTempDir(t *testing.T) string {
	t.Helper()

	base := "/tmp"
	if runtime.GOOS == "windows" {
		base = os.TempDir()
	}
	// os.MkdirTemp rather than t.TempDir: the latter follows TMPDIR,
	// which on macOS hosts exceeds the darwin sun_path limit (104 bytes)
	// once the socket path is appended.
	//nolint:usetesting
	dir, err := os.MkdirTemp(base, "gopass-agent-")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}

// startTestAgent isolates the socket directory, starts an in-process agent
// and returns a client once it answers pings. It replaces the copy-pasted
// preamble of the hello-era tests: the socket-path guard skips tests on
// environments that still exceed the platform's unix socket path limit,
// and the Eventually probe turns a failed Listen into a clear "did not
// become ready" failure instead of a misleading ping error one second
// later.
func startTestAgent(t *testing.T) *Client {
	t.Helper()

	// isolate the socket dir so the test neither requires write access to
	// the real runtime dir nor can hit a real user agent (GOPASS_HOMEDIR
	// wins over the platform runtime dir on every supported platform)
	t.Setenv("GOPASS_HOMEDIR", shortTempDir(t))

	c := NewClient()
	// a unix socket path is limited to 104 bytes (darwin) / 108 (linux);
	// skip rather than fail when an unusual environment still exceeds it
	if len(c.socketPath) > 103 {
		t.Skipf("unix socket path too long (%d bytes)", len(c.socketPath))
	}

	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

	a, err := New()
	require.NoError(t, err)
	go func() {
		_ = a.Run(ctx)
	}()
	t.Cleanup(func() {
		a.Shutdown(ctx)
	})

	require.Eventually(t, func() bool {
		return c.Ping() == nil
	}, 5*time.Second, 50*time.Millisecond, "age agent did not become ready")

	return c
}

func TestAgentHello(t *testing.T) {
	c := startTestAgent(t)

	caps, err := c.Capabilities()
	require.NoError(t, err)

	// the advertised command set is asserted against a hardcoded list so a
	// command added to the dispatch table without updating this list fails
	// here; the reverse drift (dispatched but not advertised) is
	// structurally impossible because capabilities() derives from the same
	// table.
	for _, cmd := range []string{
		"ping", "status", "identities", "decrypt",
		"lock", "lock-in", "unlock", "set-timeout", "quit",
	} {
		require.True(t, caps.Has(cmd), "missing capability %q in %q", cmd, caps.Raw())
	}
	require.False(t, caps.Has("hello"), "hello must not advertise itself")
	require.False(t, caps.Has("nonexistent"))

	v, ok := caps.Value("maxline")
	require.True(t, ok)
	require.Equal(t, "16777216", v)

	// without a stashed build version the token falls back to unknown
	v, ok = caps.Value("version")
	require.True(t, ok)
	require.NotEmpty(t, v)

	// with a stashed build version the token reports exactly that version
	buildinfo.Version = "1.2.3"
	defer func() {
		buildinfo.Version = ""
	}()
	caps, err = c.Capabilities()
	require.NoError(t, err)
	v, ok = caps.Value("version")
	require.True(t, ok)
	require.Equal(t, "1.2.3", v)

	// the optional client argument is ignored by the agent (the token list
	// is sorted, so it starts with "decrypt identities")
	resp, err := c.send("hello gopass/9.9.9")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(resp, "decrypt identities"), "unexpected response %q", resp)

	require.NoError(t, c.Quit())
}

func TestAgentCommandsAdvertised(t *testing.T) {
	c := startTestAgent(t)

	a := &Agent{}
	for cmd := range a.commandHandlers() {
		if cmd == "quit" {
			// quit shuts the agent down and removes the socket file,
			// breaking the cleanup below; its dispatch is covered by
			// the final Quit assertion.
			continue
		}
		_, err := c.send(cmd)
		if err != nil {
			// bare commands may legitimately fail on missing arguments,
			// but they must be recognized, never "unknown command".
			require.NotContains(t, err.Error(), "unknown command",
				"command %q is advertised but not implemented", cmd)
		}
	}

	require.NoError(t, c.Quit())
}

// TestAgentLockIn covers the read-only auto-lock countdown query: it reports
// -1 when no timer is armed (disabled or locked), a positive countdown while
// a timer is armed, and never resets the timer itself.
func TestAgentLockIn(t *testing.T) {
	c := startTestAgent(t)

	// no timeout configured -> no auto-lock scheduled
	secs, err := c.LockIn()
	require.NoError(t, err)
	require.Equal(t, -1, secs)

	// arm a timeout; the countdown must be positive and at most the timeout
	require.NoError(t, c.SetTimeout(60))
	secs, err = c.LockIn()
	require.NoError(t, err)
	require.Positive(t, secs)
	require.LessOrEqual(t, secs, 60)

	// polling must not reset the timer: the countdown keeps decreasing
	time.Sleep(1100 * time.Millisecond)
	secs2, err := c.LockIn()
	require.NoError(t, err)
	require.Less(t, secs2, secs)

	// disabling the timer reports -1 again
	require.NoError(t, c.SetTimeout(0))
	secs, err = c.LockIn()
	require.NoError(t, err)
	require.Equal(t, -1, secs)

	// a locked agent has no scheduled auto-lock
	require.NoError(t, c.Lock())
	secs, err = c.LockIn()
	require.NoError(t, err)
	require.Equal(t, -1, secs)

	require.NoError(t, c.Quit())
}

// TestAgentUnknownCommandKeepsConnection freezes the protocol invariant
// that an unrecognized command yields a single-line ERR, changes no state
// and does not terminate the connection (issue #3624).
func TestAgentUnknownCommandKeepsConnection(t *testing.T) {
	c := startTestAgent(t)

	conn, err := net.Dial("unix", c.socketPath)
	require.NoError(t, err)
	// fail fast with the stalled assertion named, instead of hanging until
	// the package timeout, when the agent stops answering lines
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	defer func() {
		_ = conn.Close()
	}()
	r := bufio.NewReader(conn)

	// unknown command: exactly one ERR line
	fmt.Fprintln(conn, "bogus")
	line, err := r.ReadString('\n')
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(strings.TrimSpace(line), "ERR"), "unexpected response %q", line)

	// no state change: locking still works across an unknown command
	fmt.Fprintln(conn, "lock")
	line, err = r.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "OK", strings.TrimSpace(line))

	fmt.Fprintln(conn, "bogus")
	line, err = r.ReadString('\n')
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(strings.TrimSpace(line), "ERR"))

	fmt.Fprintln(conn, "status")
	line, err = r.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "OK locked", strings.TrimSpace(line))

	// the connection keeps serving real commands, including hello
	fmt.Fprintln(conn, "ping")
	line, err = r.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "OK", strings.TrimSpace(line))

	fmt.Fprintln(conn, "hello")
	line, err = r.ReadString('\n')
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(line, "OK "), "unexpected response %q", line)

	require.NoError(t, c.Quit())
}

// TestClientCapabilitiesLegacy verifies the fallback contract: a pre-hello
// agent answering ERR to everything makes Capabilities fail, and the caller
// is expected to keep the pre-hello behaviour (judging by the error's
// existence, never its text).
func TestClientCapabilitiesLegacy(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "legacy.sock")
	l, err := net.Listen("unix", sock)
	require.NoError(t, err)
	defer func() {
		_ = l.Close()
	}()
	// client.checkSocketSecurity enforces 0600 on unix: without the chmod
	// the test would fail at the permission check instead of exercising
	// the ERR-from-legacy-agent path.
	require.NoError(t, os.Chmod(sock, 0o600))

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() {
					_ = conn.Close()
				}()
				sc := bufio.NewScanner(conn)
				for sc.Scan() {
					fmt.Fprintln(conn, "ERR unknown command")
				}
			}()
		}
	}()

	c := &Client{socketPath: sock}
	caps, err := c.Capabilities()
	require.Error(t, err)
	require.Nil(t, caps)

	// a nil Capabilities must be safe to keep and query: the nil-receiver
	// guards return the documented zero values
	require.False(t, caps.Has("ping"))
	v, ok := caps.Value("maxline")
	require.False(t, ok)
	require.Empty(t, v)
	require.Empty(t, caps.Raw())
}

func TestParseCapabilities(t *testing.T) {
	caps := parseCapabilities("ping maxline=42 future-token version=")
	require.True(t, caps.Has("ping"))
	require.True(t, caps.Has("future-token"))
	v, ok := caps.Value("maxline")
	require.True(t, ok)
	require.Equal(t, "42", v)
	v, ok = caps.Value("version")
	require.True(t, ok)
	require.Empty(t, v)
	v, ok = caps.Value("ping")
	require.True(t, ok)
	require.Empty(t, v)
	require.False(t, caps.Has("missing"))
	require.Equal(t, "ping maxline=42 future-token version=", caps.Raw())

	// empty and malformed input must not panic
	for _, in := range []string{"", "   ", "=", "=foo"} {
		require.NotNil(t, parseCapabilities(in))
	}

	// malformed tokens are skipped without affecting the rest
	caps = parseCapabilities("=foo ping")
	require.False(t, caps.Has("=foo"))
	require.True(t, caps.Has("ping"))

	// duplicate tokens: the first occurrence wins, redefinitions are
	// ignored (a bare token repeating a key=value one neither upgrades
	// nor downgrades it)
	caps = parseCapabilities("version=1.0 version=9.9")
	v, _ = caps.Value("version")
	require.Equal(t, "1.0", v)
	caps = parseCapabilities("ping ping=x")
	v, ok = caps.Value("ping")
	require.True(t, ok)
	require.Empty(t, v)
}
