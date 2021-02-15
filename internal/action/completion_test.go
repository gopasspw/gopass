package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

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
	ctx = ctxutil.WithInteractive(ctx, false)
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

	t.Run("complete foo", func(t *testing.T) {
		defer buf.Reset()

		act.Complete(gptest.CliCtx(ctx, t))
		assert.Equal(t, "foo\n", buf.String())
	})

	t.Run("bash completion", func(t *testing.T) {
		defer buf.Reset()

		assert.NoError(t, act.CompletionBash(nil))
		assert.Contains(t, buf.String(), "action.test")
	})

	t.Run("fish completion", func(t *testing.T) {
		defer buf.Reset()

		assert.NoError(t, act.CompletionFish(app))
		assert.Contains(t, buf.String(), "action.test")
		assert.Error(t, act.CompletionFish(nil))
	})

	t.Run("zsh completion", func(t *testing.T) {
		defer buf.Reset()

		assert.NoError(t, act.CompletionZSH(app))
		assert.Contains(t, buf.String(), "action.test")
		assert.Error(t, act.CompletionZSH(nil))
	})

	t.Run("openbsdksh completion", func(t *testing.T) {
		defer buf.Reset()

		assert.NoError(t, act.CompletionOpenBSDKsh(app))
		assert.Contains(t, buf.String(), "complete_gopass")
		assert.Error(t, act.CompletionOpenBSDKsh(nil))
	})
}
