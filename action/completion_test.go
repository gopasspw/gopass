package action

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	app := cli.NewApp()

	out := capture(t, func() error {
		act.Complete(nil)
		return nil
	})
	if out != "foo" {
		t.Errorf("should return 'foo' not '%s'", out)
	}

	// bash
	out = capture(t, func() error {
		return act.CompletionBash(nil)
	})
	if !strings.Contains(out, "action.test") {
		t.Errorf("should contain name of test")
	}

	// fish
	out = capture(t, func() error {
		return act.CompletionFish(nil, app)
	})
	if !strings.Contains(out, "action.test") {
		t.Errorf("should contain name of test")
	}
	assert.Error(t, act.CompletionFish(nil, nil))

	// zsh
	out = capture(t, func() error {
		return act.CompletionZSH(nil, app)
	})
	if !strings.Contains(out, "action.test") {
		t.Errorf("should contain name of test")
	}
	assert.Error(t, act.CompletionZSH(nil, nil))

	// openbsdksh
	out = capture(t, func() error {
		return act.CompletionOpenBSDKsh(nil, app)
	})
	if !strings.Contains(out, "complete_gopass") {
		t.Errorf("should contain name of test")
	}
	assert.Error(t, act.CompletionOpenBSDKsh(nil, nil))
}
