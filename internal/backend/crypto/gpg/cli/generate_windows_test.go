package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePrivateKey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	assert.NoError(t, g.CreatePrivateKeyBatch(ctx, "foo", "foo@bar.com", "bar"))
	assert.NoError(t, g.CreatePrivateKey(ctx))
	cancel()
}
