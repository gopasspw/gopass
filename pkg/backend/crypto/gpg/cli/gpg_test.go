package cli

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGPG(t *testing.T) {
	ctx := context.Background()

	var err error
	var g *GPG

	assert.Equal(t, "", g.Binary())

	g, err = New(ctx, Config{})
	require.NoError(t, err)
	assert.NotEqual(t, "", g.Binary())

	_, err = g.ListPublicKeyIDs(ctx)
	assert.NoError(t, err)

	_, err = g.ListPrivateKeyIDs(ctx)
	assert.NoError(t, err)

	_, err = g.RecipientIDs(ctx, []byte{})
	assert.Error(t, err)

	assert.NoError(t, g.Initialized(ctx))
	assert.Equal(t, "gpg", g.Name())
	assert.Equal(t, "gpg", g.Ext())
	assert.Equal(t, ".gpg-id", g.IDFile())
}

func TestDetectBinaryCandidates(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
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
