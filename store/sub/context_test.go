package sub

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
}

func TestFsckForce(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsFsckForce(ctx))
	assert.Equal(t, true, IsFsckForce(WithFsckForce(ctx, true)))
	assert.Equal(t, false, IsFsckForce(WithFsckForce(ctx, false)))
}

func TestAutoSync(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsAutoSync(ctx))
	assert.Equal(t, true, IsAutoSync(WithAutoSync(ctx, true)))
	assert.Equal(t, false, IsAutoSync(WithAutoSync(ctx, false)))
}

func TestReason(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", GetReason(ctx))
	assert.Equal(t, "foobar", GetReason(WithReason(ctx, "foobar")))
}

func TestImportFunc(t *testing.T) {
	ctx := context.Background()

	ifunc := func(context.Context, string, []string) bool {
		return true
	}
	assert.NotNil(t, GetImportFunc(ctx))
	assert.Equal(t, true, GetImportFunc(WithImportFunc(ctx, ifunc))(ctx, "", nil))
}

func TestRecipientFunc(t *testing.T) {
	ctx := context.Background()

	rfunc := func(context.Context, string, []string) ([]string, error) {
		return nil, nil
	}
	assert.NotNil(t, GetRecipientFunc(ctx))
	_, err := GetRecipientFunc(WithRecipientFunc(ctx, rfunc))(ctx, "", nil)
	assert.NoError(t, err)
}

func TestFsckFunc(t *testing.T) {
	ctx := context.Background()

	ffunc := func(context.Context, string) bool {
		return true
	}
	assert.NotNil(t, GetFsckFunc(ctx))
	assert.Equal(t, true, GetFsckFunc(WithFsckFunc(ctx, ffunc))(ctx, ""))
}

func TestCheckRecipients(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsCheckRecipients(ctx))
	assert.Equal(t, true, IsCheckRecipients(WithCheckRecipients(ctx, true)))
	assert.Equal(t, false, IsCheckRecipients(WithCheckRecipients(ctx, false)))
}
