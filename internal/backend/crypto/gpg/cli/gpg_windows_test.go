package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	_, err := g.Encrypt(ctx, []byte("foo"), nil)
	assert.NoError(t, err)
	cancel()
}

func TestDecrypt(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	_, err := g.Decrypt(ctx, []byte("foo"))
	assert.NoError(t, err)
	cancel()
}

func TestGenerateIdentity(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	assert.NoError(t, g.GenerateIdentity(ctx, "foo", "foo@bar.com", "bar"))
	cancel()
}
