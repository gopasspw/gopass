//go:build !windows
// +build !windows

package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncrypt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	_, err := g.Encrypt(ctx, []byte("foo"), nil)
	require.NoError(t, err)
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	_, err := g.Decrypt(ctx, []byte("foo"))
	require.NoError(t, err)
}

func TestGenerateIdentity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	require.NoError(t, g.GenerateIdentity(ctx, "foo", "foo@bar.com", "bar"))
}
