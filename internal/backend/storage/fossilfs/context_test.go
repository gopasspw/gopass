package fossilfs

import (
	"context"
	"testing"
)

func TestWithPathOverride(t *testing.T) {
	ctx := context.Background()
	path := "/test/path"
	ctx = withPathOverride(ctx, path)

	if val, ok := ctx.Value(ctxKeyPathOverride).(string); !ok || val != path {
		t.Errorf("Expected path %s, but got %v", path, val)
	}
}

func TestGetPathOverride(t *testing.T) {
	ctx := context.Background()
	defaultPath := "/default/path"

	// Test with no override
	if path := getPathOverride(ctx, defaultPath); path != defaultPath {
		t.Errorf("Expected default path %s, but got %s", defaultPath, path)
	}

	// Test with override
	overridePath := "/override/path"
	ctx = withPathOverride(ctx, overridePath)
	if path := getPathOverride(ctx, defaultPath); path != overridePath {
		t.Errorf("Expected override path %s, but got %s", overridePath, path)
	}
}
