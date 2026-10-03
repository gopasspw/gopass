package agent

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"testing"
	"time"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestAgent(t *testing.T) {
	ctx := t.Context()
	ctx = termio.WithPassPromptFunc(ctx, func(ctx context.Context, prompt string) (string, error) {
		return "test", nil
	})

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
