package action

import (
	"fmt"

	"github.com/urfave/cli"
)

// Version prints the gopass version
func (s *Action) Version(c *cli.Context) error {
	cli.VersionPrinter(c)

	gv := s.Store.GPGVersion()
	fmt.Printf("  GPG: %d.%d.%d\n", gv.Major, gv.Minor, gv.Patch)
	gmaj, gmin, gpa := s.Store.GitVersion()
	fmt.Printf("  Git: %d.%d.%d\n", gmaj, gmin, gpa)

	return nil
}
