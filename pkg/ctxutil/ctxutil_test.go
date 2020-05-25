package ctxutil

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestDebug(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsDebug(ctx))
	assert.Equal(t, true, IsDebug(WithDebug(ctx, true)))
	assert.Equal(t, false, IsDebug(WithDebug(ctx, false)))
}

func TestColor(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsColor(ctx))
	assert.Equal(t, true, IsColor(WithColor(ctx, true)))
	assert.Equal(t, false, IsColor(WithColor(ctx, false)))
}

func TestTerminal(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsTerminal(ctx))
	assert.Equal(t, true, IsTerminal(WithTerminal(ctx, true)))
	assert.Equal(t, false, IsTerminal(WithTerminal(ctx, false)))
}

func TestInteractive(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsInteractive(ctx))
	assert.Equal(t, true, IsInteractive(WithInteractive(ctx, true)))
	assert.Equal(t, false, IsInteractive(WithInteractive(ctx, false)))
}

func TestStdin(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsStdin(ctx))
	assert.Equal(t, true, IsStdin(WithStdin(ctx, true)))
	assert.Equal(t, false, IsStdin(WithStdin(ctx, false)))
}

func TestAskForMore(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsAskForMore(ctx))
	assert.Equal(t, true, IsAskForMore(WithAskForMore(ctx, true)))
	assert.Equal(t, false, IsAskForMore(WithAskForMore(ctx, false)))
}

func TestClipTimeout(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, 45, GetClipTimeout(ctx))
	assert.Equal(t, 30, GetClipTimeout(WithClipTimeout(ctx, 30)))
}

func TestConcurrency(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, 1, GetConcurrency(ctx))
	assert.Equal(t, 3, GetConcurrency(WithConcurrency(ctx, 3)))
}

func TestNoConfirm(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsNoConfirm(ctx))
	assert.Equal(t, true, IsNoConfirm(WithNoConfirm(ctx, true)))
	assert.Equal(t, false, IsNoConfirm(WithNoConfirm(ctx, false)))
}

func TestNoPager(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsNoPager(ctx))
	assert.Equal(t, true, IsNoPager(WithNoPager(ctx, true)))
	assert.Equal(t, false, IsNoPager(WithNoPager(ctx, false)))
}

func TestShowSafeContent(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsShowSafeContent(ctx))
	assert.Equal(t, true, IsShowSafeContent(WithShowSafeContent(ctx, true)))
	assert.Equal(t, false, IsShowSafeContent(WithShowSafeContent(ctx, false)))
}

func TestGitCommit(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsGitCommit(ctx))
	assert.Equal(t, true, IsGitCommit(WithGitCommit(ctx, true)))
	assert.Equal(t, false, IsGitCommit(WithGitCommit(ctx, false)))
}

func TestAlwaysYes(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsAlwaysYes(ctx))
	assert.Equal(t, true, IsAlwaysYes(WithAlwaysYes(ctx, true)))
	assert.Equal(t, false, IsAlwaysYes(WithAlwaysYes(ctx, false)))
}

func TestUseSymbols(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsUseSymbols(ctx))
	assert.Equal(t, true, IsUseSymbols(WithUseSymbols(ctx, true)))
	assert.Equal(t, false, IsUseSymbols(WithUseSymbols(ctx, false)))
}

func TestNoColor(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsNoColor(ctx))
	assert.Equal(t, true, IsNoColor(WithNoColor(ctx, true)))
	assert.Equal(t, false, IsNoColor(WithNoColor(ctx, false)))
}

func TestFuzzySearch(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsFuzzySearch(ctx))
	assert.Equal(t, true, IsFuzzySearch(WithFuzzySearch(ctx, true)))
	assert.Equal(t, false, IsFuzzySearch(WithFuzzySearch(ctx, false)))
}

func TestVerbose(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsVerbose(ctx))
	assert.Equal(t, true, IsVerbose(WithVerbose(ctx, true)))
	assert.Equal(t, false, IsVerbose(WithVerbose(ctx, false)))
}

func TestAutoClip(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsAutoClip(ctx))
	assert.Equal(t, true, IsAutoClip(WithAutoClip(ctx, true)))
	assert.Equal(t, false, IsAutoClip(WithAutoClip(ctx, false)))
}

func TestNotifications(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsNotifications(ctx))
	assert.Equal(t, true, IsNotifications(WithNotifications(ctx, true)))
	assert.Equal(t, false, IsNotifications(WithNotifications(ctx, false)))
}

func TestEditRecipients(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsEditRecipients(ctx))
	assert.Equal(t, true, IsEditRecipients(WithEditRecipients(ctx, true)))
	assert.Equal(t, false, IsEditRecipients(WithEditRecipients(ctx, false)))
}

func TestProgressCallback(t *testing.T) {
	ctx := context.Background()

	var foo bool
	pc := func() { foo = true }
	GetProgressCallback(WithProgressCallback(ctx, pc))()
	assert.Equal(t, true, foo)
}

func TestConfigDir(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", GetConfigDir(ctx))
	assert.Equal(t, "", GetConfigDir(WithConfigDir(ctx, "")))
}

func TestAlias(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", GetAlias(ctx))
	assert.Equal(t, "", GetAlias(WithAlias(ctx, "")))
}

func TestGitInit(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, true, IsGitInit(ctx))
	assert.Equal(t, true, IsGitInit(WithGitInit(ctx, true)))
	assert.Equal(t, false, IsGitInit(WithGitInit(ctx, false)))
}

func TestForce(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsForce(ctx))
	assert.Equal(t, true, IsForce(WithForce(ctx, true)))
	assert.Equal(t, false, IsForce(WithForce(ctx, false)))
}

func TestCommitMessage(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, "", GetCommitMessage(ctx))
	assert.Equal(t, "foo", GetCommitMessage(WithCommitMessage(ctx, "foo")))
	assert.Equal(t, "", GetCommitMessage(WithCommitMessage(ctx, "")))
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
	ctx = WithConcurrency(ctx, 5)
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
	ctx = WithEditRecipients(ctx, true)
	ctx = WithAutoClip(ctx, true)

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

	assert.Equal(t, 5, GetConcurrency(ctx))
	assert.Equal(t, true, HasConcurrency(ctx))

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

	assert.Equal(t, true, IsEditRecipients(ctx))
	assert.Equal(t, true, HasEditRecipients(ctx))

	assert.Equal(t, true, IsAutoClip(ctx))
	assert.Equal(t, true, HasAutoClip(ctx))
}

func TestGlobalFlags(t *testing.T) {
	ctx := context.Background()
	app := cli.NewApp()

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "yes",
		Usage: "yes",
	}
	assert.NoError(t, sf.Apply(fs))
	assert.NoError(t, fs.Parse([]string{"--yes"}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.Equal(t, true, IsAlwaysYes(WithGlobalFlags(c)))
}
