package action

import "github.com/urfave/cli"

// Version prints the gopass version
func (s *Action) Version(c *cli.Context) error {
	cli.VersionPrinter(c)
	return nil
}
