//go:build !linux

package action

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// secretServiceCommand returns the `gopass secret-service` command. On non-Linux
// platforms it exists only to fail with a clear message: the freedesktop.org
// Secret Service API this feature implements is Linux-only.
func (*Action) secretServiceCommand() *cli.Command {
	return &cli.Command{
		Name:  "secret-service",
		Usage: "Run a D-Bus Secret Service daemon backed by gopass (Linux only)",
		Action: func(_ context.Context, _ *cli.Command) error {
			return fmt.Errorf("secret-service is only supported on Linux")
		},
	}
}
