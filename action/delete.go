package action

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Delete a secret file with its content
func (s *Action) Delete(c *cli.Context) error {
	force := c.Bool("force")
	recursive := c.Bool("recursive")

	name := c.Args().First()
	if name == "" {
		return s.exitError(ExitUsage, nil, "Usage: %s rm name", s.Name)
	}

	key := c.Args().Get(1)

	if !force { // don't check if it's force anyway
		recStr := ""
		if recursive {
			recStr = "recursively "
		}
		if s.Store.Exists(name) && key == "" && !s.askForConfirmation(fmt.Sprintf("Are you sure you would like to %sdelete %s?", recStr, name)) {
			return nil
		}
	}

	if recursive {
		if err := s.Store.Prune(name); err != nil {
			return s.exitError(ExitUnknown, err, "failed to prune '%s': %s", name, err)
		}
		return nil
	}

	if s.Store.IsDir(name) {
		return errors.Errorf("Cannot remove '%s': Is a directory. Use 'gopass rm -r %s' to delete", name, name)
	}

	// deletes a single key from a YAML doc
	if key != "" {
		if err := s.Store.DeleteKey(name, key); err != nil {
			return s.exitError(ExitIO, err, "Can not delete key '%s' from '%s': %s", key, name, err)
		}
		return nil
	}

	if err := s.Store.Delete(name); err != nil {
		return s.exitError(ExitIO, err, "Can not delete '%s': %s", name, err)
	}
	return nil
}
