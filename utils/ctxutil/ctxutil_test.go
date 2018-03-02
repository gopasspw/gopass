package ctxutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestStdin(t *testing.T) {
	ctx := context.Background()

	if IsStdin(ctx) {
		t.Errorf("Vanilla ctx should not have Stdin")
	}
	if !IsStdin(WithStdin(ctx, true)) {
		t.Errorf("Should have stdin flag")
	}
	if IsStdin(WithStdin(ctx, false)) {
		t.Errorf("Should not have Stdin flag")
	}
}

func TestAskForMore(t *testing.T) {
	ctx := context.Background()

	if IsAskForMore(ctx) {
		t.Errorf("Vanilla ctx should not have AskForMore")
	}
	if !IsAskForMore(WithAskForMore(ctx, true)) {
		t.Errorf("Should have AskForMore flag")
	}
	if IsAskForMore(WithAskForMore(ctx, false)) {
		t.Errorf("Should not have AskForMore flag")
	}
}

func TestClipTimeout(t *testing.T) {
	ctx := context.Background()

	if GetClipTimeout(ctx) != 45 {
		t.Errorf("Vanilla ctx should have ClipTimeout 45")
	}
	if GetClipTimeout(WithClipTimeout(ctx, 30)) != 30 {
		t.Errorf("ClipTimeout should be 30")
	}
}

func TestNoConfirm(t *testing.T) {
	ctx := context.Background()

	if IsNoConfirm(ctx) {
		t.Errorf("Vanilla ctx should not have NoConfirm")
	}
	if !IsNoConfirm(WithNoConfirm(ctx, true)) {
		t.Errorf("Should have NoConfirm flag")
	}
	if IsNoConfirm(WithNoConfirm(ctx, false)) {
		t.Errorf("Should not have NoConfirm flag")
	}
}

func TestNoPager(t *testing.T) {
	ctx := context.Background()

	if IsNoPager(ctx) {
		t.Errorf("Vanilla ctx should not have NoPager")
	}
	if !IsNoPager(WithNoPager(ctx, true)) {
		t.Errorf("Should have NoPager flag")
	}
	if IsNoPager(WithNoPager(ctx, false)) {
		t.Errorf("Should not have NoPager flag")
	}
}

func TestShowSafeContent(t *testing.T) {
	ctx := context.Background()

	if IsShowSafeContent(ctx) {
		t.Errorf("Vanilla ctx should not have ShowSafeContent")
	}
	if !IsShowSafeContent(WithShowSafeContent(ctx, true)) {
		t.Errorf("Should have ShowSafeContent flag")
	}
	if IsShowSafeContent(WithShowSafeContent(ctx, false)) {
		t.Errorf("Should not have ShowSafeContent flag")
	}
}

func TestGitCommit(t *testing.T) {
	ctx := context.Background()

	if !IsGitCommit(ctx) {
		t.Errorf("Vanilla ctx should have GitCommit")
	}
	if !IsGitCommit(WithGitCommit(ctx, true)) {
		t.Errorf("Should have GitCommit flag")
	}
	if IsGitCommit(WithGitCommit(ctx, false)) {
		t.Errorf("Should not have GitCommit flag")
	}
}

func TestUseSymbols(t *testing.T) {
	ctx := context.Background()

	if IsUseSymbols(ctx) {
		t.Errorf("Vanilla ctx should not have UseSymbols")
	}
	if !IsUseSymbols(WithUseSymbols(ctx, true)) {
		t.Errorf("Should have UseSymbols flag")
	}
	if IsUseSymbols(WithUseSymbols(ctx, false)) {
		t.Errorf("Should not have UseSymbols flag")
	}
}

func TestNoColor(t *testing.T) {
	ctx := context.Background()

	if IsNoColor(ctx) {
		t.Errorf("Vanilla ctx should not have NoColor")
	}
	if !IsNoColor(WithNoColor(ctx, true)) {
		t.Errorf("Should have NoColor flag")
	}
	if IsNoColor(WithNoColor(ctx, false)) {
		t.Errorf("Should not have NoColor flag")
	}
}

func TestAlwaysYes(t *testing.T) {
	ctx := context.Background()

	if IsAlwaysYes(ctx) {
		t.Errorf("Vanilla ctx should not have AlwaysYes")
	}
	if !IsAlwaysYes(WithAlwaysYes(ctx, true)) {
		t.Errorf("Should have AlwaysYes flag")
	}
	if IsAlwaysYes(WithAlwaysYes(ctx, false)) {
		t.Errorf("Should not have AlwaysYes flag")
	}
}

func TestFuzzySearch(t *testing.T) {
	ctx := context.Background()

	if !IsFuzzySearch(ctx) {
		t.Errorf("Vanilla ctx should have FuzzySearch")
	}
	if !IsFuzzySearch(WithFuzzySearch(ctx, true)) {
		t.Errorf("Should have FuzzySearch flag")
	}
	if IsFuzzySearch(WithFuzzySearch(ctx, false)) {
		t.Errorf("Should not have FuzzySearch flag")
	}
}

func TestVerbose(t *testing.T) {
	ctx := context.Background()

	if IsVerbose(ctx) {
		t.Errorf("Vanilla ctx should not have Verbose")
	}
	if !IsVerbose(WithVerbose(ctx, true)) {
		t.Errorf("Should have Verbose flag")
	}
	if IsVerbose(WithVerbose(ctx, false)) {
		t.Errorf("Should not have Verbose flag")
	}
}

func TestNotifications(t *testing.T) {
	ctx := context.Background()

	if !IsNotifications(ctx) {
		t.Errorf("Vanilla ctx should not have Notifications")
	}
	if !IsNotifications(WithNotifications(ctx, true)) {
		t.Errorf("Should have Notifications flag")
	}
	if IsNotifications(WithNotifications(ctx, false)) {
		t.Errorf("Should not have Notifications flag")
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
	ctx = WithNotifications(ctx, true)

	assert.Equal(t, true, IsDebug(ctx))
	assert.Equal(t, true, HasDebug(ctx))

	assert.Equal(t, false, IsColor(ctx))
	assert.Equal(t, true, HasColor(ctx))

	assert.Equal(t, false, IsTerminal(ctx))
	assert.Equal(t, true, HasTerminal(ctx))

	assert.Equal(t, false, IsInteractive(ctx))
	assert.Equal(t, true, HasInteractive(ctx))

	assert.Equal(t, true, IsStdin(ctx))
	assert.Equal(t, true, HasStdin(ctx))

	assert.Equal(t, true, IsAskForMore(ctx))
	assert.Equal(t, true, HasAskForMore(ctx))

	assert.Equal(t, 10, GetClipTimeout(ctx))
	assert.Equal(t, true, HasClipTimeout(ctx))

	assert.Equal(t, true, IsNoConfirm(ctx))
	assert.Equal(t, true, HasNoConfirm(ctx))

	assert.Equal(t, true, IsNoPager(ctx))
	assert.Equal(t, true, HasNoPager(ctx))

	assert.Equal(t, true, IsShowSafeContent(ctx))
	assert.Equal(t, true, HasShowSafeContent(ctx))

	assert.Equal(t, false, IsGitCommit(ctx))
	assert.Equal(t, true, HasGitCommit(ctx))

	assert.Equal(t, false, IsUseSymbols(ctx))
	assert.Equal(t, true, HasUseSymbols(ctx))

	assert.Equal(t, true, IsAlwaysYes(ctx))
	assert.Equal(t, true, HasAlwaysYes(ctx))

	assert.Equal(t, true, IsNoColor(ctx))
	assert.Equal(t, true, HasNoColor(ctx))

	assert.Equal(t, false, IsFuzzySearch(ctx))
	assert.Equal(t, true, HasFuzzySearch(ctx))

	assert.Equal(t, true, IsVerbose(ctx))
	assert.Equal(t, true, HasVerbose(ctx))

	assert.Equal(t, true, IsNotifications(ctx))
	assert.Equal(t, true, HasNotifications(ctx))
}
