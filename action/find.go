package action

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

// Find a string in the secret file's name
func (s *Action) Find(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("Usage: gopass find arg")
	}

	l, err := s.Store.List(0)
	if err != nil {
		return err
	}
	needle := strings.ToLower(c.Args().First())
	for _, value := range l {
		if strings.Contains(strings.ToLower(value), needle) {
			fmt.Println(value)
		}
	}

	return nil
}
