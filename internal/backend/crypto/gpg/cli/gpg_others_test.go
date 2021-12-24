//go:build !windows
// +build !windows

package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	_, err := g.Encrypt(ctx, []byte("foo"), nil)
	assert.NoError(t, err)
}

func TestDecrypt(t *testing.T) {
	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	_, err := g.Decrypt(ctx, []byte("foo"))
	assert.NoError(t, err)
}

func TestGenerateIdentity(t *testing.T) {
	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	assert.NoError(t, g.GenerateIdentity(ctx, "foo", "foo@bar.com", "bar"))
}
