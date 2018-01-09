package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportPublicKey(t *testing.T) {
	ctx := context.Background()
	g, err := New(ctx, Config{})
	assert.NoError(t, err)

	_, err = g.ExportPublicKey(ctx, "foobar")
	assert.Error(t, err)
}
