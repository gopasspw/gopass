package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBinaryCandidates(t *testing.T) {
	bins, err := detectBinaryCandidates("foobar")
	assert.NoError(t, err)
	assert.Contains(t, bins, "C:\\Program Files (x86)\\GnuPG\\bin\\gpg.exe")
}

func TestEncrypt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	_, err := g.Encrypt(ctx, []byte("foo"), nil)
	assert.NoError(t, err)
	cancel()
}

func TestDecrypt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	_, err := g.Decrypt(ctx, []byte("foo"))
	assert.NoError(t, err)
	cancel()
}
