package action

import (
	"context"
	"testing"
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

func TestWithForce(t *testing.T) {
	ctx := context.Background()

	if IsForce(ctx) {
		t.Errorf("Should be false")
	}

	if !IsForce(WithForce(ctx, true)) {
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

	if IsPrintQR(ctx) {
		t.Errorf("Should be false")
	}

	if !IsPrintQR(WithPrintQR(ctx, true)) {
		t.Errorf("Should be true")
	}
}
