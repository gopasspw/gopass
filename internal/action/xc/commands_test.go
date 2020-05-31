// +build xc

package xc

import (
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func testCommand(t *testing.T, cmd *cli.Command) {
	if len(cmd.Subcommands) < 1 {
		assert.NotNil(t, cmd.Action, cmd.Name)
	}
	assert.NotEmpty(t, cmd.Usage, cmd.Name)
	assert.NotEmpty(t, cmd.Description, cmd.Name)
	for _, flag := range cmd.Flags {
		switch v := flag.(type) {
		case *cli.StringFlag:
			assert.NotContains(t, v.Name, ",")
			assert.NotEmpty(t, v.Usage, flag.Names())
		case *cli.BoolFlag:
			assert.NotContains(t, v.Name, ",")
			assert.NotEmpty(t, v.Usage, flag.Names())
		}
	}
	for _, scmd := range cmd.Subcommands {
		testCommand(t, scmd)
	}
}

func TestCommands(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	for _, cmd := range GetCommands() {
		testCommand(t, cmd)
	}
}
