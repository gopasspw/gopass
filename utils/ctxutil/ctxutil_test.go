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
	ctx = WithFuzzySearch(ctx, false)
	ctx = WithVerbose(ctx, true)

	if !IsDebug(ctx) {
		t.Errorf("Debug should be true")
	}
	if !HasDebug(ctx) {
		t.Errorf("Should have Debug")
	}
	if IsColor(ctx) {
		t.Errorf("Color should be false")
	}
	if !HasColor(ctx) {
		t.Errorf("Should have Color")
	}
	if IsTerminal(ctx) {
		t.Errorf("Terminal should be false")
	}
	if !HasTerminal(ctx) {
		t.Errorf("Should have Terminal")
	}
	if IsInteractive(ctx) {
		t.Errorf("IsInteractive should be false")
	}
	if !HasInteractive(ctx) {
		t.Errorf("Should have Interactive")
	}
	if !IsStdin(ctx) {
		t.Errorf("IsStdin should be true")
	}
	if !HasStdin(ctx) {
		t.Errorf("Should have Stdin")
	}
	if !IsAskForMore(ctx) {
		t.Errorf("Ask for more should be true")
	}
	if !HasAskForMore(ctx) {
		t.Errorf("Should have AskForMore")
	}
	if GetClipTimeout(ctx) != 10 {
		t.Errorf("Clip timeout should be 10")
	}
	if !HasClipTimeout(ctx) {
		t.Errorf("Should have ClipTimeout")
	}
	if !IsNoConfirm(ctx) {
		t.Errorf("NoConfirm should be true")
	}
	if !HasNoConfirm(ctx) {
		t.Errorf("Should have NoConfirm")
	}
	if !IsNoPager(ctx) {
		t.Errorf("NoPager should be true")
	}
	if !HasNoPager(ctx) {
		t.Errorf("Should have NoPager")
	}
	if !IsShowSafeContent(ctx) {
		t.Errorf("ShowSafeContexnt should be true")
	}
	if !HasShowSafeContent(ctx) {
		t.Errorf("Should have ShowSafeContent")
	}
	if IsGitCommit(ctx) {
		t.Errorf("Git commit should be false")
	}
	if !HasGitCommit(ctx) {
		t.Errorf("Shoud have GitCommit")
	}
	if IsUseSymbols(ctx) {
		t.Errorf("UseSymbols should be false")
	}
	if !HasUseSymbols(ctx) {
		t.Errorf("Should have UseSymbols")
	}
	if !IsAlwaysYes(ctx) {
		t.Errorf("Always yes should be true")
	}
	if !HasAlwaysYes(ctx) {
		t.Errorf("Should have AlwaysYes")
	}
	if !IsNoColor(ctx) {
		t.Errorf("NoColor should be true")
	}
	if !HasNoColor(ctx) {
		t.Errorf("Should have NoColor")
	}
	if IsFuzzySearch(ctx) {
		t.Errorf("FuzzySearch should be false")
	}
	if !IsVerbose(ctx) {
		t.Errorf("Verbose should be true")
	}
}
