package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportPublicKey(t *testing.T) {
	ctx := context.Background()
	g, err := New(ctx, Config{})
	require.NoError(t, err)

	_, err = g.ExportPublicKey(ctx, "foobar")
	assert.Error(t, err)
}
