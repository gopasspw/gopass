// +build xc

package xc

import (
	"github.com/urfave/cli/v2"
)

// GetCommands returns the CLI commands exported for the XC crypto backend
func GetCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "xc",
			Usage:  "Experimental Crypto",
			Hidden: true,
			Description: "" +
				"These subcommands are used to control and test the experimental crypto" +
				"implementation.",
			Subcommands: []*cli.Command{
				{
					Name:   "list-private-keys",
					Action: ListPrivateKeys,
				},
				{
					Name:   "list-public-keys",
					Action: ListPublicKeys,
				},
				{
					Name:   "generate",
					Action: GenerateKeypair,
				},
				{
					Name:   "export",
					Action: ExportPublicKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "id",
						},
						&cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name:   "import",
					Action: ImportPublicKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "id",
						},
						&cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name:   "export-private-key",
					Action: ExportPrivateKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "id",
						},
						&cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name:   "import-private-key",
					Action: ImportPrivateKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "id",
						},
						&cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name:   "remove",
					Action: RemoveKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "id",
						},
					},
				},
				{
					Name:   "encrypt",
					Action: EncryptFile,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "file",
						},
						&cli.StringSliceFlag{
							Name: "recipients",
						},
						&cli.BoolFlag{
							Name: "stream",
						},
					},
				},
				{
					Name:   "decrypt",
					Action: DecryptFile,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "file",
						},
						&cli.BoolFlag{
							Name: "stream",
						},
					},
				},
			},
		},
	}
}
