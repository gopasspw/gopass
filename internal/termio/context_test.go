package termio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassPromptFunc(t *testing.T) {
	ctx := context.Background()

	assert.False(t, HasPassPromptFunc(ctx))
	assert.NotNil(t, GetPassPromptFunc(ctx))

	ctx = WithPassPromptFunc(ctx, func(context.Context, string) (string, error) {
		return "test", nil
	})
	assert.True(t, HasPassPromptFunc(ctx))
	assert.NotNil(t, GetPassPromptFunc(ctx))
	sv, err := GetPassPromptFunc(ctx)(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, "test", sv)
}
