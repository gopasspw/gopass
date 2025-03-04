package age

import (
	"context"
	"testing"
)

func TestWithOnlyNative(t *testing.T) {
	ctx := context.Background()
	ctx = WithOnlyNative(ctx, true)

	val := ctx.Value(ctxKeyOnlyNative)
	if val == nil {
		t.Errorf("Expected value to be set, got nil")
	}

	boolVal, ok := val.(bool)
	if !ok {
		t.Errorf("Expected value to be of type bool, got %T", val)
	}

	if !boolVal {
		t.Errorf("Expected value to be true, got false")
	}
}

func TestIsOnlyNative(t *testing.T) {
	ctx := context.Background()

	// Test default value
	if IsOnlyNative(ctx) {
		t.Errorf("Expected default value to be false, got true")
	}

	// Test set value
	ctx = WithOnlyNative(ctx, true)
	if !IsOnlyNative(ctx) {
		t.Errorf("Expected value to be true, got false")
	}

	// Test reset value
	ctx = WithOnlyNative(ctx, false)
	if IsOnlyNative(ctx) {
		t.Errorf("Expected value to be false, got true")
	}
}
