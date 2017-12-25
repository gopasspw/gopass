package action

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

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

	// zsh
	out = capture(t, func() error {
		return act.CompletionZSH(nil, app)
	})
	if !strings.Contains(out, "action.test") {
		t.Errorf("should contain name of test")
	}
}
