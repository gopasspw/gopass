package action

import (
	"fmt"

	"github.com/justwatchcom/gopass/store"
	"github.com/urfave/cli"
)

// Delete a secret file with its content
func (s *Action) Delete(c *cli.Context) error {
	force := c.Bool("force")
	recursive := c.Bool("recursive")

	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("provide a secret name")
	}

	found, err := s.Store.Exists(name)
	if err != nil && err != store.ErrNotFound {
		return fmt.Errorf("failed to see if %s exists", name)
	}

	if !force { // don't check if it's force anyway
		recStr := ""
		if recursive {
			recStr = "recursively "
		}
		if found && !askForConfirmation(fmt.Sprintf("Are you sure you would like to %sdelete %s?", recStr, name)) {
			return nil
		}
	}

	if recursive {
		return s.Store.Prune(name)
	}

	if s.Store.IsDir(name) {
		return fmt.Errorf("Cannot remove '%s': Is a directory. Use 'gopass rm -r %s' to delete", name, name)
	}

	return s.Store.Delete(name)
}
