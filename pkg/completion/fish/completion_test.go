package fish

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

type unknownFlag struct{}

func (u unknownFlag) String() string {
	return ""
}

func (u unknownFlag) Apply(*flag.FlagSet) {}
func (u unknownFlag) GetName() string {
	return ""
}

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
		{"print", "Print", "short", ""},
		{"", "Print", "long", ""},
		{"print, p", "Print", "foo", ""},
	} {
		assert.Equal(t, tc.Out, formatFlag(tc.Name, tc.Usage, tc.Typ))
	}
}

func TestGetCompletion(t *testing.T) {
	app := cli.NewApp()
	sv, err := GetCompletion(app)
	require.NoError(t, err)
	assert.Contains(t, sv, "#!/usr/bin/env fish")

	fishTemplate = "{{.unexported}}"
	sv, err = GetCompletion(app)
	assert.Error(t, err)
	assert.Contains(t, sv, "")

	fishTemplate = "{{}}"
	sv, err = GetCompletion(app)
	assert.Error(t, err)
	assert.Contains(t, sv, "")
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
		require.NoError(t, err)
		assert.Equal(t, "", sv)

		sv, err = formatFlagFunc("long")(flag)
		require.NoError(t, err)
		assert.Equal(t, "foo", sv)

		sv, err = formatFlagFunc("usage")(flag)
		require.NoError(t, err)
		assert.Equal(t, "bar", sv)
	}

	sv, err := formatFlagFunc("short")(unknownFlag{})
	assert.Error(t, err)
	assert.Equal(t, "", sv)

	sv, err = formatFlagFunc("long")(unknownFlag{})
	assert.Error(t, err)
	assert.Equal(t, "", sv)

	sv, err = formatFlagFunc("usage")(unknownFlag{})
	assert.Error(t, err)
	assert.Equal(t, "", sv)
}
