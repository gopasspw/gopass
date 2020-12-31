package out

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefix(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", Prefix(ctx))

	ctx = AddPrefix(ctx, "[foo] ")
	assert.Equal(t, "[foo] ", Prefix(ctx))

	ctx = AddPrefix(ctx, "[bar] ")
	assert.Equal(t, "[foo] [bar] ", Prefix(ctx))

	ctx = AddPrefix(ctx, "")
	assert.Equal(t, "[foo] [bar] ", Prefix(ctx))
}

func TestHidden(t *testing.T) {
	ctx := context.Background()

	assert.False(t, IsHidden(ctx))
	assert.True(t, IsHidden(WithHidden(ctx, true)))
}

func TestNewline(t *testing.T) {
	ctx := context.Background()

	assert.True(t, HasNewline(ctx))
	assert.False(t, HasNewline(WithNewline(ctx, false)))
}
