package fish

import (
	"testing"

	"github.com/urfave/cli"
)

func TestFormatFlag(t *testing.T) {
	for _, tc := range []struct {
		Name  string
		Usage string
		Typ   string
		Out   string
	}{
		{"print, p", "Print", "short", "p"},
		{"print, p", "Print", "long", "print"},
		{"print, p", "Print", "usage", "Print"},
		{"print, p", "Print", "foo", ""},
	} {
		out := formatFlag(tc.Name, tc.Usage, tc.Typ)
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
