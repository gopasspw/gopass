package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
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
	app.Commands = []cli.Command{
		{
			Name:    "test",
			Aliases: []string{"foo", "bar"},
		},
	}

	act.Complete(ctx, nil)
	assert.Equal(t, "foo\n", buf.String())
	buf.Reset()

	// bash
	assert.NoError(t, act.CompletionBash(nil))
	assert.Contains(t, buf.String(), "action.test")
	buf.Reset()

	// fish
	assert.NoError(t, act.CompletionFish(nil, app))
	assert.Contains(t, buf.String(), "action.test")
	assert.Error(t, act.CompletionFish(nil, nil))
	buf.Reset()

	// zsh
	assert.NoError(t, act.CompletionZSH(nil, app))
	assert.Contains(t, buf.String(), "action.test")
	assert.Error(t, act.CompletionZSH(nil, nil))
	buf.Reset()

	// openbsdksh
	assert.NoError(t, act.CompletionOpenBSDKsh(nil, app))
	assert.Contains(t, buf.String(), "complete_gopass")
	assert.Error(t, act.CompletionOpenBSDKsh(nil, nil))
	buf.Reset()
}

func TestFilterCompletionList(t *testing.T) {
	for _, tc := range []struct {
		name   string
		in     []string
		needle string
		out    []string
	}{
		{
			name:   "empty",
			in:     []string{"foo", "bar", "misc/baz", "misc/zab"},
			needle: "",
			out:    []string{"bar", "foo", "misc"},
		},
		{
			name:   "misc/",
			in:     []string{"foo", "bar", "misc/baz", "misc/zab", "misc/zab/abc"},
			needle: "misc/",
			out:    []string{"misc/baz", "misc/zab"},
		},
		{
			name:   "misc",
			in:     []string{"foo", "bar", "misc/baz", "misc/zab", "misc/zab/abc"},
			needle: "misc",
			out:    []string{"misc/"},
		},
		{
			name:   "web",
			in:     []string{"foo", "bar", "misc/baz", "webmaster/foo", "websites/bar"},
			needle: "web",
			out:    []string{"webmaster", "websites"},
		},
	} {
		assert.Equal(t, tc.out, filterCompletionList(tc.in, tc.needle), tc.name)
	}
}
