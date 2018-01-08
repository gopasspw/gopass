package action

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestBashEscape(t *testing.T) {
	expected := `a\\<\\>\\|\\\\and\\ sometimes\\?\\*\\(\\)\\&\\;\\#`
	if escaped := bashEscape(`a<>|\and sometimes?*()&;#`); escaped != expected {
		t.Errorf("Expected %q, but got %q", expected, escaped)
	}
}

func TestComplete(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	app := cli.NewApp()

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
