package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Git runs git commands inside the store or mounts
func (s *Action) Git(c *cli.Context) error {
	store := c.String("store")
	return s.Store.Git(store, c.Args()...)
}

// GitInit initializes a git repo
func (s *Action) GitInit(c *cli.Context) error {
	store := c.String("store")
	sk := c.String("sign-key")

	return s.gitInit(store, sk)
}

func (s *Action) gitInit(store, sk string) error {
	if sk == "" {
		s, err := askForPrivateKey(color.CyanString("Please select a key for signing Git Commits"))
		if err == nil {
			sk = s
		}
	}

	if err := s.Store.GitInit(store, sk); err != nil {
		return err
	}
	fmt.Fprintln(color.Output, color.GreenString("Git initialized"))
	return nil
}
