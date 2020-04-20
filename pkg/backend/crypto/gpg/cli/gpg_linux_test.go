package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBinaryCandidates(t *testing.T) {
	bins, err := detectBinaryCandidates("foobar")
	assert.NoError(t, err)
	assert.Equal(t, []string{"gpg2", "gpg1", "gpg", "foobar"}, bins)
}

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
