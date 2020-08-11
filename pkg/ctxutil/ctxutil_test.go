package ctxutil

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

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

func TestConfirm(t *testing.T) {
	ctx := context.Background()

	assert.Equal(t, false, IsConfirm(ctx))
	assert.Equal(t, true, IsConfirm(WithConfirm(ctx, true)))
	assert.Equal(t, false, IsConfirm(WithConfirm(ctx, false)))
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

func TestProgressCallback(t *testing.T) {
	ctx := context.Background()

	var foo bool
	pc := func() { foo = true }
	GetProgressCallback(WithProgressCallback(ctx, pc))()
	assert.Equal(t, true, foo)
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
	ctx = WithColor(ctx, false)
	ctx = WithTerminal(ctx, false)
	ctx = WithInteractive(ctx, false)
	ctx = WithStdin(ctx, true)
	ctx = WithClipTimeout(ctx, 10)
	ctx = WithConcurrency(ctx, 5)
	ctx = WithConfirm(ctx, true)
	ctx = WithNoPager(ctx, true)
	ctx = WithShowSafeContent(ctx, true)
	ctx = WithGitCommit(ctx, false)
	ctx = WithAlwaysYes(ctx, true)
	ctx = WithNoColor(ctx, true)
	ctx = WithFuzzySearch(ctx, false)
	ctx = WithVerbose(ctx, true)
	ctx = WithNotifications(ctx, true)
	ctx = WithAutoClip(ctx, true)
	ctx = WithExportKeys(ctx, false)
	ctx = WithEmail(ctx, "foo@bar.com")
	ctx = WithUsername(ctx, "foo")
	ctx = WithNoNetwork(ctx, true)
	ctx = WithCommitMessage(ctx, "foobar")
	ctx = WithForce(ctx, true)
	ctx = WithGitInit(ctx, false)

	assert.Equal(t, false, IsColor(ctx))
	assert.Equal(t, true, HasColor(ctx))

	assert.Equal(t, false, IsTerminal(ctx))
	assert.Equal(t, true, HasTerminal(ctx))

	assert.Equal(t, false, IsInteractive(ctx))
	assert.Equal(t, true, HasInteractive(ctx))

	assert.Equal(t, true, IsStdin(ctx))
	assert.Equal(t, true, HasStdin(ctx))

	assert.Equal(t, 10, GetClipTimeout(ctx))
	assert.Equal(t, true, HasClipTimeout(ctx))

	assert.Equal(t, 5, GetConcurrency(ctx))
	assert.Equal(t, true, HasConcurrency(ctx))

	assert.Equal(t, true, IsConfirm(ctx))
	assert.Equal(t, true, HasConfirm(ctx))

	assert.Equal(t, true, IsNoPager(ctx))
	assert.Equal(t, true, HasNoPager(ctx))

	assert.Equal(t, true, IsShowSafeContent(ctx))
	assert.Equal(t, true, HasShowSafeContent(ctx))

	assert.Equal(t, false, IsGitCommit(ctx))
	assert.Equal(t, true, HasGitCommit(ctx))

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

	assert.Equal(t, true, IsAutoClip(ctx))
	assert.Equal(t, true, HasAutoClip(ctx))

	assert.Equal(t, false, IsExportKeys(ctx))
	assert.Equal(t, true, HasExportKeys(ctx))

	assert.Equal(t, "foo@bar.com", GetEmail(ctx))
	assert.Equal(t, "foo", GetUsername(ctx))

	assert.Equal(t, true, IsNoNetwork(ctx))
	assert.Equal(t, true, HasNoNetwork(ctx))

	assert.Equal(t, "foobar", GetCommitMessage(ctx))
	assert.Equal(t, true, HasCommitMessage(ctx))

	assert.Equal(t, true, IsForce(ctx))
	assert.Equal(t, true, HasForce(ctx))

	assert.Equal(t, false, IsGitInit(ctx))
	assert.Equal(t, true, HasGitInit(ctx))
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

func TestImportFunc(t *testing.T) {
	ctx := context.Background()

	ifunc := func(context.Context, string, []string) bool {
		return true
	}
	assert.NotNil(t, GetImportFunc(ctx))
	assert.Equal(t, true, GetImportFunc(WithImportFunc(ctx, ifunc))(ctx, "", nil))
	assert.Equal(t, true, HasImportFunc(WithImportFunc(ctx, ifunc)))
	assert.Equal(t, true, GetImportFunc(WithImportFunc(ctx, nil))(ctx, "", nil))
}
