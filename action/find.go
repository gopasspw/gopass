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

	l, err := s.Store.List()
	if err != nil {
		return err
	}
	for _, value := range l {
		if strings.Contains(value, c.Args().First()) {
			fmt.Println(value)
		}
	}

	return nil
}
