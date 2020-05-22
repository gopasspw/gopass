package cli

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBinaryCandidates(t *testing.T) {
	bins, err := detectBinaryCandidates("foobar")
	assert.NoError(t, err)
	// the install locations differ depending on :
	// - chocolatey install path prefix
	// - 64bit/32bit windows
	var stripped []string
	for _, bin := range bins {
		stripped = append(stripped, filepath.Base(bin))
	}
	assert.Contains(t, stripped, "gpg.exe")
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
