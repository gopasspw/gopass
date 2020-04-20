package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	g := &GPG{}
	g.binary = "rundll32"

	assert.NoError(t, g.ImportPublicKey(ctx, []byte("foobar")))

	g.binary = ""
	assert.Error(t, g.ImportPublicKey(ctx, []byte("foobar")))
	cancel()
}
