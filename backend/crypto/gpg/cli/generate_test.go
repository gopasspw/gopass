package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePrivateKey(t *testing.T) {
	ctx := context.Background()

	g := &GPG{}
	g.binary = "true"

	assert.NoError(t, g.CreatePrivateKeyBatch(ctx, "foo", "foo@bar.com", "bar"))
	assert.NoError(t, g.CreatePrivateKey(ctx))
}
