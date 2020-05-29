package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestBashEscape(t *testing.T) {
	expected := `a\\<\\>\\|\\\\and\\ sometimes\\?\\*\\(\\)\\&\\;\\#`
	if escaped := bashEscape(`a<>|\and sometimes?*()&;#`); escaped != expected {
		t.Errorf("Expected %q, but got %q", expected, escaped)
	}
}

func TestComplete(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name:    "test",
			Aliases: []string{"foo", "bar"},
		},
	}

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	act.Complete(c)
	assert.Equal(t, "foo\n", buf.String())
	buf.Reset()

	// bash
	assert.NoError(t, act.CompletionBash(nil))
	assert.Contains(t, buf.String(), "action.test")
	buf.Reset()

	// fish
	assert.NoError(t, act.CompletionFish(app))
	assert.Contains(t, buf.String(), "action.test")
	assert.Error(t, act.CompletionFish(nil))
	buf.Reset()

	// zsh
	assert.NoError(t, act.CompletionZSH(app))
	assert.Contains(t, buf.String(), "action.test")
	assert.Error(t, act.CompletionZSH(nil))
	buf.Reset()

	// openbsdksh
	assert.NoError(t, act.CompletionOpenBSDKsh(app))
	assert.Contains(t, buf.String(), "complete_gopass")
	assert.Error(t, act.CompletionOpenBSDKsh(nil))
	buf.Reset()
}
