package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateIdentity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	assert.NoError(t, g.GenerateIdentity(ctx, "foo", "foo@bar.com", "bar"))
	cancel()
}
