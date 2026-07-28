package agent

import (
	"bytes"
	"context"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/require"
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
