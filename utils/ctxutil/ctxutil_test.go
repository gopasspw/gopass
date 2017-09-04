package ctxutil

import (
	"context"
	"testing"
)

func TestDebug(t *testing.T) {
	if IsDebug(context.Background()) {
		t.Errorf("Vanilla ctx should not be debug")
	}
	if !IsDebug(WithDebug(context.Background(), true)) {
		t.Errorf("Should have debug flag")
	}
	if IsDebug(WithDebug(context.Background(), false)) {
		t.Errorf("Should not have debug flag")
	}
}

func TestColor(t *testing.T) {
	if !IsColor(context.Background()) {
		t.Errorf("Vanilla ctx should have color")
	}
	if !IsColor(WithColor(context.Background(), true)) {
		t.Errorf("Should have color flag")
	}
	if IsColor(WithColor(context.Background(), false)) {
		t.Errorf("Should not have color flag")
	}
}

func TestTerminal(t *testing.T) {
	if !IsTerminal(context.Background()) {
		t.Errorf("Vanilla ctx should be terminal")
	}
	if !IsTerminal(WithTerminal(context.Background(), true)) {
		t.Errorf("Should have terminal flag")
	}
	if IsTerminal(WithTerminal(context.Background(), false)) {
		t.Errorf("Should not have terminal flag")
	}
}

func TestInteractive(t *testing.T) {
	if !IsInteractive(context.Background()) {
		t.Errorf("Vanilla ctx should be interactive")
	}
	if !IsInteractive(WithInteractive(context.Background(), true)) {
		t.Errorf("Should have interactive flag")
	}
	if IsInteractive(WithInteractive(context.Background(), false)) {
		t.Errorf("Should not have interactive flag")
	}
}

func TestComposite(t *testing.T) {
	ctx := context.Background()
	ctx = WithColor(ctx, true)
	ctx = WithTerminal(ctx, false)
	ctx = WithInteractive(ctx, false)
	ctx = WithColor(ctx, false)

	if IsColor(ctx) {
		t.Errorf("Color should be false")
	}
}
