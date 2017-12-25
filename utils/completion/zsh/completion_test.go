package zsh

import (
	"testing"

	"github.com/urfave/cli"
)

func TestFormatFlag(t *testing.T) {
	for _, tc := range []struct {
		Name  string
		Usage string
		Out   string
	}{
		{"print, p", "Print", "--print[Print]"},
	} {
		out := formatFlag(tc.Name, tc.Usage)
		if out != tc.Out {
			t.Errorf("'%s' != '%s'", out, tc.Out)
		}
	}
}

func TestGetCompletion(t *testing.T) {
	app := cli.NewApp()
	sv, err := GetCompletion(app)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	t.Logf("Output: %s", sv)
}
