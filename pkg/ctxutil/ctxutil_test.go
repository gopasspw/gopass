package ctxutil

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestTerminal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.True(t, IsTerminal(ctx))
	assert.True(t, IsTerminal(WithTerminal(ctx, true)))
	assert.False(t, IsTerminal(WithTerminal(ctx, false)))
}

func TestInteractive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.True(t, IsInteractive(ctx))
	assert.True(t, IsInteractive(WithInteractive(ctx, true)))
	assert.False(t, IsInteractive(WithInteractive(ctx, false)))
}

func TestStdin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.False(t, IsStdin(ctx))
	assert.True(t, IsStdin(WithStdin(ctx, true)))
	assert.False(t, IsStdin(WithStdin(ctx, false)))
}

func TestGitCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.True(t, IsGitCommit(ctx))
	assert.True(t, IsGitCommit(WithGitCommit(ctx, true)))
	assert.False(t, IsGitCommit(WithGitCommit(ctx, false)))
}

func TestAlwaysYes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.False(t, IsAlwaysYes(ctx))
	assert.True(t, IsAlwaysYes(WithAlwaysYes(ctx, true)))
	assert.False(t, IsAlwaysYes(WithAlwaysYes(ctx, false)))
}

func TestProgressCallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var foo bool

	pc := func() { foo = true }

	GetProgressCallback(WithProgressCallback(ctx, pc))()
	assert.True(t, foo)
}

func TestAlias(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.Equal(t, "", GetAlias(ctx))
	assert.Equal(t, "", GetAlias(WithAlias(ctx, "")))
}

func TestGitInit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.True(t, IsGitInit(ctx))
	assert.True(t, IsGitInit(WithGitInit(ctx, true)))
	assert.False(t, IsGitInit(WithGitInit(ctx, false)))
}

func TestForce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.False(t, IsForce(ctx))
	assert.True(t, IsForce(WithForce(ctx, true)))
	assert.False(t, IsForce(WithForce(ctx, false)))
}

func TestCommitMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.Equal(t, "", GetCommitMessage(ctx))
	assert.Equal(t, "foo", GetCommitMessage(WithCommitMessage(ctx, "foo")))
	assert.Equal(t, "", GetCommitMessage(WithCommitMessage(ctx, "")))
}

func TestCommitMessageBody(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx2 := AddToCommitMessageBody(AddToCommitMessageBody(WithCommitMessage(ctx, "foo"), "bar"), "baz")
	assert.Equal(t, "foo\n\nbar\nbaz", GetCommitMessageFull(ctx2))
	assert.Equal(t, "foo", GetCommitMessage(ctx2))
	assert.Equal(t, "bar\nbaz", GetCommitMessageBody(ctx2))
	ctx2 = AddToCommitMessageBody(AddToCommitMessageBody(ctx, "bar"), "baz")
	assert.Equal(t, "", GetCommitMessage(ctx2))
	assert.Equal(t, "bar\nbaz", GetCommitMessageFull(ctx2))
	assert.Equal(t, "bar\nbaz", GetCommitMessageBody(ctx2))
	ctx2 = WithCommitMessage(ctx, "foo")
	assert.Equal(t, "foo", GetCommitMessage(ctx2))
	assert.Equal(t, "foo", GetCommitMessageFull(ctx2))
	assert.Equal(t, "", GetCommitMessageBody(ctx2))
}

func TestComposite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = WithTerminal(ctx, false)
	ctx = WithInteractive(ctx, false)
	ctx = WithStdin(ctx, true)
	ctx = WithGitCommit(ctx, false)
	ctx = WithAlwaysYes(ctx, true)
	ctx = WithEmail(ctx, "foo@bar.com")
	ctx = WithUsername(ctx, "foo")
	ctx = WithNoNetwork(ctx, true)
	ctx = WithCommitMessage(ctx, "foobar")
	ctx = WithForce(ctx, true)
	ctx = WithGitInit(ctx, false)

	assert.False(t, IsTerminal(ctx))
	assert.True(t, HasTerminal(ctx))

	assert.False(t, IsInteractive(ctx))
	assert.True(t, HasInteractive(ctx))

	assert.True(t, IsStdin(ctx))
	assert.True(t, HasStdin(ctx))

	assert.False(t, IsGitCommit(ctx))
	assert.True(t, HasGitCommit(ctx))

	assert.True(t, IsAlwaysYes(ctx))
	assert.True(t, HasAlwaysYes(ctx))

	assert.Equal(t, "foo@bar.com", GetEmail(ctx))
	assert.Equal(t, "foo", GetUsername(ctx))

	assert.True(t, IsNoNetwork(ctx))
	assert.True(t, HasNoNetwork(ctx))

	assert.Equal(t, "foobar", GetCommitMessage(ctx))
	assert.True(t, HasCommitMessage(ctx))

	assert.True(t, IsForce(ctx))
	assert.True(t, HasForce(ctx))

	assert.False(t, IsGitInit(ctx))
	assert.True(t, HasGitInit(ctx))
}

func TestGlobalFlags(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := cli.NewApp()

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "yes",
		Usage: "yes",
	}
	require.NoError(t, sf.Apply(fs))
	require.NoError(t, fs.Parse([]string{"--yes"}))
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	assert.True(t, IsAlwaysYes(WithGlobalFlags(c)))
}

func TestImportFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ifunc := func(context.Context, string, []string) bool {
		return true
	}

	assert.NotNil(t, GetImportFunc(ctx))
	assert.True(t, GetImportFunc(WithImportFunc(ctx, ifunc))(ctx, "", nil))
	assert.True(t, HasImportFunc(WithImportFunc(ctx, ifunc)))
	assert.True(t, GetImportFunc(WithImportFunc(ctx, nil))(ctx, "", nil))
}

func TestHidden(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.False(t, IsHidden(ctx))
	assert.True(t, IsHidden(WithHidden(ctx, true)))
}
