package action

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

// List all secrets as a tree
func (s *Action) List(c *cli.Context) error {
	filter := c.Args().First()
	flat := c.Bool("flat")
	stripPrefix := c.Bool("strip-prefix")
	limit := c.Int("limit")

	l, err := s.Store.Tree()
	if err != nil {
		return err
	}

	if filter == "" {
		if flat {
			for _, e := range l.List(limit) {
				fmt.Println(e)
			}
			return nil
		}
		fmt.Println(l.Format(limit))
		return nil
	}

	if subtree := l.FindFolder(filter); subtree != nil {
		subtree.Root = true
		subtree.Name = filter
		if flat {
			sep := "/"
			if strings.HasSuffix(filter, "/") {
				sep = ""
			}
			for _, e := range subtree.List(limit) {
				if stripPrefix {
					fmt.Println(e)
					continue
				}
				fmt.Println(filter + sep + e)
			}
			return nil
		}
		fmt.Println(subtree.Format(limit))
		return nil
	}

	return nil
}
