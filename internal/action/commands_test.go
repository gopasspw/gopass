package action

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func testCommand(t *testing.T, cmd *cli.Command) {
	t.Helper()

	if len(cmd.Subcommands) < 1 {
		assert.NotNil(t, cmd.Action, cmd.Name)
	}

	assert.NotEmpty(t, cmd.Usage)
	assert.NotEmpty(t, cmd.Description)

	for _, flag := range cmd.Flags {
		switch v := flag.(type) {
		case *cli.StringFlag:
			assert.NotContains(t, v.Name, ",")
			assert.NotEmpty(t, v.Usage)
		case *cli.BoolFlag:
			assert.NotContains(t, v.Name, ",")
			assert.NotEmpty(t, v.Usage)
		}
	}

	for _, scmd := range cmd.Subcommands {
		testCommand(t, scmd)
	}
}

func TestCommands(t *testing.T) {
	t.Parallel()

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	for _, cmd := range act.GetCommands() {
		cmd := cmd
		t.Run(cmd.Name, func(t *testing.T) {
			t.Parallel()

			testCommand(t, cmd)
		})
	}
}
