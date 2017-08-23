package action

import (
	"fmt"

	"github.com/urfave/cli"
)

// Copy the contents of a file to another one
func (s *Action) Copy(c *cli.Context) error {
	force := c.Bool("force")

	if len(c.Args()) != 2 {
		return s.exitError(ExitUsage, nil, "Usage: %s cp old-path new-path", s.Name)
	}

	from := c.Args()[0]
	to := c.Args()[1]

	if !s.Store.Exists(from) {
		return s.exitError(ExitNotFound, nil, "%s does not exist", from)
	}

	if !force {
		if s.Store.Exists(to) && !s.askForConfirmation(fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return s.exitError(ExitAborted, nil, "not overwriting your current secret")
		}
	}

	if err := s.Store.Copy(from, to); err != nil {
		return s.exitError(ExitIO, err, "failed to copy from '%s' to '%s'", from, to)
	}

	return nil
}
