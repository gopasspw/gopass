package binary

import (
	"github.com/urfave/cli/v2"
)

type initializer interface {
	Initialized(*cli.Context) error
	Complete(*cli.Context)
}

// GetCommands returns the CLI commands exported for the create commands
func GetCommands(i initializer, store storer) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "binary",
			Usage: "Assist with Binary/Base64 content",
			Description: "" +
				"These commands directly convert binary files from/to base64 encoding.",
			Aliases: []string{"bin"},
			Hidden:  true,
			Subcommands: []*cli.Command{
				{
					Name:  "cat",
					Usage: "Print content of a secret to stdout, or insert from stdin",
					Description: "" +
						"This command is similar to the way cat works on the command line. " +
						"It can either be used to retrieve the decoded content of a secret " +
						"similar to 'cat file' or vice versa to encode the content from STDIN " +
						"to a secret.",
					Before: i.Initialized,
					Action: func(c *cli.Context) error {
						return Cat(c, store)
					},
					BashComplete: i.Complete,
				},
				{
					Name:  "sum",
					Usage: "Compute the SHA256 checksum",
					Description: "" +
						"This command decodes an Base64 encoded secret and computes the SHA256 checksum " +
						"over the decoded data. This is useful to verify the integrity of an " +
						"inserted secret.",
					Aliases: []string{"sha", "sha256"},
					Before:  i.Initialized,
					Action: func(c *cli.Context) error {
						return Sum(c, store)
					},
					BashComplete: i.Complete,
				},
				{
					Name:  "copy",
					Usage: "Copy files from or to the password store",
					Description: "" +
						"This command either reads a file from the filesystem and writes the " +
						"encoded and encrypted version in the store or it decrypts and decodes " +
						"a secret and writes the result to a file. Either source or destination " +
						"must be a file and the other one a secret. If you want the source to " +
						"be securely removed after copying, use 'gopass binary move'",
					Before:  i.Initialized,
					Aliases: []string{"cp"},
					Action: func(c *cli.Context) error {
						return Copy(c, store)
					},
					BashComplete: i.Complete,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "force",
							Aliases: []string{"f"},
							Usage:   "Force to move the secret and overwrite existing one",
						},
					},
				},
				{
					Name:  "move",
					Usage: "Move files from or to the password store",
					Description: "" +
						"This command either reads a file from the filesystem and writes the " +
						"encoded and encrypted version in the store or it decrypts and decodes " +
						"a secret and writes the result to a file. Either source or destination " +
						"must be a file and the other one a secret. The source will be wiped " +
						"from disk or from the store after it has been copied successfully " +
						"and validated. If you don't want the source to be removed use " +
						"'gopass binary copy'",
					Before:  i.Initialized,
					Aliases: []string{"mv"},
					Action: func(c *cli.Context) error {
						return Move(c, store)
					},
					BashComplete: i.Complete,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "force",
							Aliases: []string{"f"},
							Usage:   "Force to move the secret and overwrite existing one",
						},
					},
				},
			},
		},
	}
}
