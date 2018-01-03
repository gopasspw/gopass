package fish

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)
	assert.Contains(t, sv, "#!/usr/bin/env fish")
}

func TestFormatflagFunc(t *testing.T) {
	for _, flag := range []cli.Flag{
		cli.BoolFlag{Name: "foo", Usage: "bar"},
		cli.Float64Flag{Name: "foo", Usage: "bar"},
		cli.GenericFlag{Name: "foo", Usage: "bar"},
		cli.Int64Flag{Name: "foo", Usage: "bar"},
		cli.Int64SliceFlag{Name: "foo", Usage: "bar"},
		cli.IntFlag{Name: "foo", Usage: "bar"},
		cli.IntSliceFlag{Name: "foo", Usage: "bar"},
		cli.StringFlag{Name: "foo", Usage: "bar"},
		cli.StringSliceFlag{Name: "foo", Usage: "bar"},
		cli.Uint64Flag{Name: "foo", Usage: "bar"},
		cli.UintFlag{Name: "foo", Usage: "bar"},
	} {
		sv, err := formatFlagFunc("short")(flag)
		assert.NoError(t, err)
		assert.Equal(t, "", sv)

		sv, err = formatFlagFunc("long")(flag)
		assert.NoError(t, err)
		assert.Equal(t, "foo", sv)

		sv, err = formatFlagFunc("usage")(flag)
		assert.NoError(t, err)
		assert.Equal(t, "bar", sv)
	}
}
