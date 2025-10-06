package leaf

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestFsckCheck(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	assert.False(t, IsFsckCheck(ctx))
	assert.True(t, IsFsckCheck(WithFsckCheck(ctx, true)))
	assert.False(t, IsFsckCheck(WithFsckCheck(ctx, false)))
	assert.True(t, HasFsckCheck(WithFsckCheck(ctx, true)))
}

func TestFsckForce(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	assert.False(t, IsFsckForce(ctx))
	assert.True(t, IsFsckForce(WithFsckForce(ctx, true)))
	assert.False(t, IsFsckForce(WithFsckForce(ctx, false)))
	assert.True(t, HasFsckForce(WithFsckForce(ctx, true)))
}

func TestFsckFunc(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	ffunc := func(context.Context, string) bool {
		return true
	}
	assert.NotNil(t, GetFsckFunc(ctx))
	assert.True(t, GetFsckFunc(ctx)(ctx, ""))
	assert.True(t, GetFsckFunc(WithFsckFunc(ctx, ffunc))(ctx, ""))
	assert.True(t, HasFsckFunc(WithFsckFunc(ctx, ffunc)))
}
