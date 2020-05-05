package xc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAskPass(t *testing.T) {
	ctx := context.Background()

	ap := newAskPass()
	ap.testing = true

	assert.NoError(t, ap.Ping(ctx))
	assert.NoError(t, ap.Remove(ctx, "foo"))

	pw, err := ap.Passphrase(ctx, "foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "", pw)
}
