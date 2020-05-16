package create

import (
	"github.com/urfave/cli/v2"
)

type initializer interface {
	Initialized(*cli.Context) error
}

// GetCommands returns the CLI commands exported for the create commands
func GetCommands(i initializer, store storer) []*cli.Command {
	return []*cli.Command{
		{
			Name:    "create",
			Aliases: []string{"new"},
			Usage:   "Easy creation of new secrets",
			Description: "" +
				"This command starts a wizard to aid in creation of new secrets.",
			Before: i.Initialized,
			Action: func(c *cli.Context) error {
				return Create(c, store)
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "store",
					Aliases: []string{"s"},
					Usage:   "Which store to use",
				},
			},
		},
	}
}
