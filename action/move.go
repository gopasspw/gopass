package action

import (
	"fmt"

	"github.com/urfave/cli"
)

// Move the content from one secret to another
func (s *Action) Move(c *cli.Context) error {
	force := c.Bool("force")

	if len(c.Args()) != 2 {
		return s.exitError(ExitUsage, nil, "Usage: %s mv old-path new-path", s.Name)
	}

	from := c.Args()[0]
	to := c.Args()[1]

	if !force {
		if s.Store.Exists(to) && !s.askForConfirmation(fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return s.exitError(ExitAborted, nil, "not overwriting your current secret")
		}
	}

	if err := s.Store.Move(from, to); err != nil {
		return s.exitError(ExitUnknown, err, "%s", err)
	}

	return nil
}
