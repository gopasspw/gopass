package out

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPrefix(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())

	assert.Equal(t, "", Prefix(ctx))

	ctx = AddPrefix(ctx, "[foo] ")
	assert.Equal(t, "[foo] ", Prefix(ctx))

	ctx = AddPrefix(ctx, "[bar] ")
	assert.Equal(t, "[foo] [bar] ", Prefix(ctx))

	ctx = AddPrefix(ctx, "")
	assert.Equal(t, "[foo] [bar] ", Prefix(ctx))
}

func TestNewline(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())

	assert.True(t, HasNewline(ctx))
	assert.False(t, HasNewline(WithNewline(ctx, false)))
}
