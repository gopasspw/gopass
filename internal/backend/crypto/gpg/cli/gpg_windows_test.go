package cli

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

func TestEncrypt(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(config.NewContextInMemory())

	g := &GPG{}
	g.binary = "rundll32"

	_, err := g.Encrypt(ctx, []byte("foo"), nil)
	require.NoError(t, err)
	cancel()
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(config.NewContextInMemory())

	g := &GPG{}
	g.binary = "rundll32"

	_, err := g.Decrypt(ctx, []byte("foo"))
	require.NoError(t, err)
	cancel()
}

func TestGenerateIdentity(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(config.NewContextInMemory())

	g := &GPG{}
	g.binary = "rundll32"

	require.NoError(t, g.GenerateIdentity(ctx, "foo", "foo@bar.com", "bar"))
	cancel()
}
