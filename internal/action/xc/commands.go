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
					Name:        "list-private-keys",
					Action:      ListPrivateKeys,
					Usage:       "List private Keys",
					Description: "List private Keys",
				},
				{
					Name:        "list-public-keys",
					Action:      ListPublicKeys,
					Usage:       "List public keys",
					Description: "List private Keys",
				},
				{
					Name:        "generate",
					Action:      GenerateKeypair,
					Usage:       "Generate new keypair",
					Description: "Generate new keypair",
				},
				{
					Name:        "export",
					Usage:       "Export a public key",
					Description: "Export a public key",
					Action:      ExportPublicKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "id",
							Usage: "Key ID",
						},
						&cli.StringFlag{
							Name:  "file",
							Usage: "Filename",
						},
					},
				},
				{
					Name:        "import",
					Usage:       "Import a public key",
					Description: "Import a public key",
					Action:      ImportPublicKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "id",
							Usage: "Key ID",
						},
						&cli.StringFlag{
							Name:  "file",
							Usage: "Filename",
						},
					},
				},
				{
					Name:        "export-private-key",
					Usage:       "Export a private key",
					Description: "Export a private key",
					Action:      ExportPrivateKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "id",
							Usage: "Key ID",
						},
						&cli.StringFlag{
							Name:  "file",
							Usage: "Filename",
						},
					},
				},
				{
					Name:        "import-private-key",
					Usage:       "Import a private key",
					Description: "Import a private key",
					Action:      ImportPrivateKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "id",
							Usage: "Key ID",
						},
						&cli.StringFlag{
							Name:  "file",
							Usage: "Filename",
						},
					},
				},
				{
					Name:        "remove",
					Usage:       "Remove a public key",
					Description: "Remove a public key",
					Action:      RemoveKey,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "id",
							Usage: "Key ID",
						},
					},
				},
				{
					Name:        "encrypt",
					Usage:       "Encrypt a file",
					Description: "Encrypt a file for recipients",
					Action:      EncryptFile,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "file",
							Usage: "Filename",
						},
						&cli.StringSliceFlag{
							Name:  "recipients",
							Usage: "List of recipients",
						},
						&cli.BoolFlag{
							Name:  "stream",
							Usage: "Encrypt in streaming mode",
						},
					},
				},
				{
					Name:        "decrypt",
					Usage:       "Decrypt a file",
					Description: "Decrypt a file encrypted for you",
					Action:      DecryptFile,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "file",
							Usage: "Filename",
						},
						&cli.BoolFlag{
							Name:  "stream",
							Usage: "Decrypt in streaming mode",
						},
					},
				},
			},
		},
	}
}
