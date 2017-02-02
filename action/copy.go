package action

import (
	"fmt"

	"github.com/urfave/cli"
)

// Copy the contents of a file to another one
func (s *Action) Copy(c *cli.Context) error {
	force := c.Bool("force")

	if len(c.Args()) != 2 {
		return fmt.Errorf("Usage: gopass cp old-path new-path")
	}

	from := c.Args()[0]
	to := c.Args()[1]

	exists, err := s.Store.Exists(from)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("%s doesn't exists", from)
	}

	if !force {
		exists, err := s.Store.Exists(to)
		if err != nil {
			return err
		}
		if exists && !askForConfirmation(fmt.Sprintf("%s already exists. Overwrite it?", to)) {
			return fmt.Errorf("not overwriting your current secret")
		}
	}

	if err := s.Store.Copy(from, to); err != nil {
		return err
	}

	return nil
}
