package leaf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsckCheck(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsFsckCheck(ctx))
	assert.Equal(t, true, IsFsckCheck(WithFsckCheck(ctx, true)))
	assert.Equal(t, false, IsFsckCheck(WithFsckCheck(ctx, false)))
	assert.Equal(t, true, HasFsckCheck(WithFsckCheck(ctx, true)))
}

func TestFsckForce(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsFsckForce(ctx))
	assert.Equal(t, true, IsFsckForce(WithFsckForce(ctx, true)))
	assert.Equal(t, false, IsFsckForce(WithFsckForce(ctx, false)))
	assert.Equal(t, true, HasFsckForce(WithFsckForce(ctx, true)))
}

func TestFsckFunc(t *testing.T) {
	ctx := context.Background()

	ffunc := func(context.Context, string) bool {
		return true
	}
	assert.NotNil(t, GetFsckFunc(ctx))
	assert.Equal(t, true, GetFsckFunc(ctx)(ctx, ""))
	assert.Equal(t, true, GetFsckFunc(WithFsckFunc(ctx, ffunc))(ctx, ""))
	assert.Equal(t, true, HasFsckFunc(WithFsckFunc(ctx, ffunc)))
}

func TestCheckRecipients(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsCheckRecipients(ctx))
	assert.Equal(t, true, IsCheckRecipients(WithCheckRecipients(ctx, true)))
	assert.Equal(t, false, IsCheckRecipients(WithCheckRecipients(ctx, false)))
	assert.Equal(t, true, HasCheckRecipients(WithCheckRecipients(ctx, true)))
}
