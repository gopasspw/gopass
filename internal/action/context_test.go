package action

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithClip(t *testing.T) {
	ctx := context.Background()

	if IsClip(ctx) {
		t.Errorf("Should be false")
	}

	if !IsClip(WithClip(ctx, true)) {
		t.Errorf("Should be true")
	}
}

func TestWithPasswordOnly(t *testing.T) {
	ctx := context.Background()

	if IsPasswordOnly(ctx) {
		t.Errorf("Should be false")
	}

	if !IsPasswordOnly(WithPasswordOnly(ctx, true)) {
		t.Errorf("Should be true")
	}
}

func TestWithPrintQR(t *testing.T) {
	ctx := context.Background()

	assert.False(t, IsPrintQR(ctx))
	assert.True(t, IsPrintQR(WithPrintQR(ctx, true)))
}

func TestWithRevision(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", GetRevision(ctx))
	assert.Equal(t, "foo", GetRevision(WithRevision(ctx, "foo")))
	assert.False(t, HasRevision(ctx))
	assert.True(t, HasRevision(WithRevision(ctx, "foo")))
}

func TestWithKey(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", GetKey(ctx))
	assert.Equal(t, "foo", GetKey(WithKey(ctx, "foo")))
}

func TestWithOnlyClip(t *testing.T) {
	ctx := context.Background()

	assert.False(t, IsOnlyClip(ctx))
	assert.True(t, IsOnlyClip(WithOnlyClip(ctx, true)))
}

func TestWithAlsoClip(t *testing.T) {
	ctx := context.Background()

	assert.False(t, IsAlsoClip(ctx))
	assert.True(t, IsAlsoClip(WithAlsoClip(ctx, true)))
}
