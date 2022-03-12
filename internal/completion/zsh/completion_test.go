package zsh

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

type unknownFlag struct{}

func (u *unknownFlag) String() string {
	return ""
}

func (u *unknownFlag) Apply(*flag.FlagSet) error {
	return nil
}

func (u *unknownFlag) GetName() string {
	return ""
}

func (u *unknownFlag) IsSet() bool {
	return true
}

func (u *unknownFlag) Names() []string {
	return []string{}
}

func TestFormatFlag(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		usage string
		out   string
	}{
		{"print, p", "Print", "--print[Print]"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.out, formatFlag(tc.name, tc.usage))
		})
	}
}

func TestGetCompletion(t *testing.T) {
	t.Parallel()

	app := cli.NewApp()
	sv, err := GetCompletion(app)
	require.NoError(t, err)
	assert.Contains(t, sv, "#compdef zsh.test")

	zshTemplate = "{{.unexported}}"
	sv, err = GetCompletion(app)
	assert.Error(t, err)
	assert.Contains(t, sv, "")

	zshTemplate = "{{}}"
	sv, err = GetCompletion(app)
	assert.Error(t, err)
	assert.Contains(t, sv, "")
}

func TestFormatflagFunc(t *testing.T) {
	t.Parallel()

	ff := formatFlagFunc()
	for _, flag := range []cli.Flag{
		&cli.BoolFlag{Name: "foo", Usage: "bar"},
		&cli.Float64Flag{Name: "foo", Usage: "bar"},
		&cli.GenericFlag{Name: "foo", Usage: "bar"},
		&cli.Int64Flag{Name: "foo", Usage: "bar"},
		&cli.Int64SliceFlag{Name: "foo", Usage: "bar"},
		&cli.IntFlag{Name: "foo", Usage: "bar"},
		&cli.IntSliceFlag{Name: "foo", Usage: "bar"},
		&cli.StringFlag{Name: "foo", Usage: "bar"},
		&cli.StringSliceFlag{Name: "foo", Usage: "bar"},
		&cli.Uint64Flag{Name: "foo", Usage: "bar"},
		&cli.UintFlag{Name: "foo", Usage: "bar"},
	} {
		sv, err := ff(flag)
		require.NoError(t, err)
		assert.Equal(t, "--foo[bar]", sv)
	}

	sv, err := ff(&unknownFlag{})
	assert.Error(t, err)
	assert.Equal(t, "", sv)
}
