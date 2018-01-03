package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPacket(t *testing.T) {
	m := splitPacket(":pubkey enc packet: version 3, algo 16, keyid 6780DF473C7A71D3")
	val, found := m["keyid"]
	if !found {
		t.Errorf("Failed to parse/lookup keyid")
	}
	if val != "6780DF473C7A71D3" {
		t.Errorf("Failed to get keyid")
	}
}

func TestGPG(t *testing.T) {
	ctx := context.Background()
	g, err := New(ctx, Config{})
	assert.NoError(t, err)
	assert.NotEqual(t, "", g.Binary())

	_, err = g.ListPublicKeys(ctx)
	assert.NoError(t, err)

	_, err = g.ListPrivateKeys(ctx)
	assert.NoError(t, err)

	_, err = g.GetRecipients(ctx, "nothing")
	assert.Error(t, err)
}
