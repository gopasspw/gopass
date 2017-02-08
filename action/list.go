package action

import (
	"fmt"

	"github.com/urfave/cli"
)

// List all secrets as a tree
func (s *Action) List(c *cli.Context) error {
	raw := c.Bool("raw")
	filter := c.Args().First()

	// Don't show a tree only new lines
	if raw {
		//TODO(metalmatze): Support filtering
		s.Complete(c)
		return nil
	}

	l, err := s.Store.Tree()
	if err != nil {
		return err
	}

	if filter == "" {
		fmt.Println(l.Format())
		return nil
	}

	if subtree := l.FindFolder(filter); subtree != nil {
		subtree.Root = true
		subtree.Name = filter
		fmt.Println(subtree.Format())
		return nil
	}

	return nil
}
