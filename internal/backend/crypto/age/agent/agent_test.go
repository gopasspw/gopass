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
	id, err := age.NewScryptIdentity("test")
	require.NoError(t, err)
	recip, err := age.NewScryptRecipient("test")
	require.NoError(t, err)

	plaintext := []byte("hello world")
	ciphertext, err := age.Encrypt(bytes.NewReader(plaintext), recip)
	require.NoError(t, err)

	require.NoError(t, c.SendIdentities(id.String()))

	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// test lock
	require.NoError(t, c.Lock())

	// test quit
	require.NoError(t, c.Quit())
}
