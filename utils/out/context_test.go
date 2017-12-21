package out

import (
	"context"
	"testing"
)

func TestPrefix(t *testing.T) {
	ctx := context.Background()

	if pfx := Prefix(ctx); pfx != "" {
		t.Errorf("non-empty prefix: %s", pfx)
	}

	ctx = AddPrefix(ctx, "[foo] ")
	if pfx := Prefix(ctx); pfx != "[foo] " {
		t.Errorf("invalid prefix: %s", pfx)
	}

	ctx = AddPrefix(ctx, "[bar] ")
	if pfx := Prefix(ctx); pfx != "[foo] [bar] " {
		t.Errorf("invalid prefix: %s", pfx)
	}

	ctx = AddPrefix(ctx, "")
	if pfx := Prefix(ctx); pfx != "[foo] [bar] " {
		t.Errorf("invalid prefix: %s", pfx)
	}
}

func TestHidden(t *testing.T) {
	ctx := context.Background()

	if IsHidden(ctx) {
		t.Errorf("hidden should be false")
	}

	if !IsHidden(WithHidden(ctx, true)) {
		t.Errorf("hidden should be true")
	}
}

func TestNewline(t *testing.T) {
	ctx := context.Background()

	if !HasNewline(ctx) {
		t.Errorf("Newline should be true")
	}

	if HasNewline(WithNewline(ctx, false)) {
		t.Errorf("Newline should be false")
	}
}
