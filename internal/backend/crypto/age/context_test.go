package age

import (
	"context"
	"testing"
)

func TestWithOnlyNative(t *testing.T) {
	ctx := t.Context()
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
	ctx := t.Context()

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

func TestAgentLauncherContext(t *testing.T) {
	ctx := t.Context()
	if l := GetAgentLauncher(ctx); l != nil {
		t.Errorf("expected nil launcher by default, got a non-nil func")
	}

	called := false
	fn := func(context.Context) error {
		called = true

		return nil
	}
	ctx = WithAgentLauncher(ctx, fn)

	got := GetAgentLauncher(ctx)
	if got == nil {
		t.Fatalf("expected non-nil launcher after WithAgentLauncher")
	}
	if err := got(ctx); err != nil {
		t.Fatalf("launcher returned error: %v", err)
	}
	if !called {
		t.Errorf("expected launcher to be called")
	}
}
