package ctxutil

import (
	"context"
	"testing"
)

func TestDebug(t *testing.T) {
	ctx := context.Background()

	if IsDebug(ctx) {
		t.Errorf("Vanilla ctx should not be debug")
	}
	if !IsDebug(WithDebug(ctx, true)) {
		t.Errorf("Should have debug flag")
	}
	if IsDebug(WithDebug(ctx, false)) {
		t.Errorf("Should not have debug flag")
	}
}

func TestColor(t *testing.T) {
	ctx := context.Background()

	if !IsColor(ctx) {
		t.Errorf("Vanilla ctx should have color")
	}
	if !IsColor(WithColor(ctx, true)) {
		t.Errorf("Should have color flag")
	}
	if IsColor(WithColor(ctx, false)) {
		t.Errorf("Should not have color flag")
	}
}

func TestTerminal(t *testing.T) {
	ctx := context.Background()

	if !IsTerminal(ctx) {
		t.Errorf("Vanilla ctx should be terminal")
	}
	if !IsTerminal(WithTerminal(ctx, true)) {
		t.Errorf("Should have terminal flag")
	}
	if IsTerminal(WithTerminal(ctx, false)) {
		t.Errorf("Should not have terminal flag")
	}
}

func TestInteractive(t *testing.T) {
	ctx := context.Background()

	if !IsInteractive(ctx) {
		t.Errorf("Vanilla ctx should be interactive")
	}
	if !IsInteractive(WithInteractive(ctx, true)) {
		t.Errorf("Should have interactive flag")
	}
	if IsInteractive(WithInteractive(ctx, false)) {
		t.Errorf("Should not have interactive flag")
	}
}

func TestComposite(t *testing.T) {
	ctx := context.Background()
	ctx = WithDebug(ctx, true)
	ctx = WithColor(ctx, false)
	ctx = WithTerminal(ctx, false)
	ctx = WithInteractive(ctx, false)
	ctx = WithStdin(ctx, true)
	ctx = WithAskForMore(ctx, true)
	ctx = WithClipTimeout(ctx, 10)
	ctx = WithNoConfirm(ctx, true)
	ctx = WithNoPager(ctx, true)
	ctx = WithShowSafeContent(ctx, true)
	ctx = WithGitCommit(ctx, false)
	ctx = WithUseSymbols(ctx, false)
	ctx = WithAlwaysYes(ctx, true)
	ctx = WithNoColor(ctx, true)

	if !IsDebug(ctx) {
		t.Errorf("Debug should be true")
	}
	if IsColor(ctx) {
		t.Errorf("Color should be false")
	}
	if IsTerminal(ctx) {
		t.Errorf("Termiunal should be false")
	}
	if IsInteractive(ctx) {
		t.Errorf("IsInteractive should be false")
	}
	if !IsStdin(ctx) {
		t.Errorf("IsStdin should be true")
	}
	if !IsAskForMore(ctx) {
		t.Errorf("Ask for more should be true")
	}
	if GetClipTimeout(ctx) != 10 {
		t.Errorf("Clip timeout should be 10")
	}
	if !IsNoConfirm(ctx) {
		t.Errorf("NoConfirm should be true")
	}
	if !IsNoPager(ctx) {
		t.Errorf("NoPager should be true")
	}
	if !IsShowSafeContent(ctx) {
		t.Errorf("ShowSafeContexnt should be true")
	}
	if IsGitCommit(ctx) {
		t.Errorf("Git commit should be false")
	}
	if IsUseSymbols(ctx) {
		t.Errorf("UseSymbols should be false")
	}
	if !IsAlwaysYes(ctx) {
		t.Errorf("Always yes should be true")
	}
}
