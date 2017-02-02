package action

import (
	"fmt"

	"github.com/urfave/cli"
)

// Move the content from one secret to another
func (s *Action) Move(c *cli.Context) error {
	force := c.Bool("force")

	if len(c.Args()) != 2 {
		return fmt.Errorf("Usage: gopass mv old-path new-path")
	}

	from := c.Args()[0]
	to := c.Args()[1]

	if !force {
		exists, err := s.Store.Exists(to)
		if err != nil {
			return err
		}
		if exists && !askForConfirmation(fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return fmt.Errorf("not overwriting your current secret")
		}
	}

	if err := s.Store.Move(from, to); err != nil {
		return err
	}

	return nil
}
