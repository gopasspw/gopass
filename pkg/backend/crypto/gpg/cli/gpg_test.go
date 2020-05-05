package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGPG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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
