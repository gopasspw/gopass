package termio

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassPromptFunc(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())

	assert.False(t, HasPassPromptFunc(ctx))
	assert.NotNil(t, GetPassPromptFunc(ctx))

	ctx = WithPassPromptFunc(ctx, func(context.Context, string) (string, error) {
		return "test", nil
	})
	assert.True(t, HasPassPromptFunc(ctx))
	assert.NotNil(t, GetPassPromptFunc(ctx))
	sv, err := GetPassPromptFunc(ctx)(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, "test", sv)
}
