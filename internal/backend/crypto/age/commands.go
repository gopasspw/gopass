package age

import (
	"fmt"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

func (l loader) Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   name,
			Hidden: true,
			Usage:  "age commands",
			Description: "" +
				"Built-in commands for the age backend.\n" +
				"These allow limited interactions with the gopass specific age identities.",
			Subcommands: []*cli.Command{
				{
					Name:  "identities",
					Usage: "List identities",
					Description: "" +
						"List identities",
					Action: func(c *cli.Context) error {
						ctx := ctxutil.WithGlobalFlags(c)
						a, err := New(ctx)
						if err != nil {
							return exit.Error(exit.Unknown, err, "failed to create age backend")
						}

						ids, err := a.IdentityRecipients(ctx)
						if err != nil {
							return exit.Error(exit.Unknown, err, "failed to get age identities")
						}

						if len(ids) < 1 {
							out.Notice(ctx, "No identities found")
						}

						for _, id := range recipientsToBech32(ids) {
							out.Printf(ctx, id)
						}

						return nil
					},
					Subcommands: []*cli.Command{
						{
							Name:  "add",
							Usage: "Add an identity",
							Description: "" +
								"Add an identity",
							Action: func(c *cli.Context) error {
								ctx := ctxutil.WithGlobalFlags(c)
								a, err := New(ctx)
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to create age backend")
								}

								if err := a.GenerateIdentity(ctx, "", "", ""); err != nil {
									return exit.Error(exit.Unknown, err, "failed to generate age identity")
								}

								return nil
							},
						},
						{
							Name:  "remove",
							Usage: "Remove an identity",
							Description: "" +
								"Remove an identity",
							Action: func(c *cli.Context) error {
								ctx := ctxutil.WithGlobalFlags(c)
								a, err := New(ctx)
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to create age backend")
								}
								victim := c.Args().First()

								ids, _ := a.Identities(ctx)
								newIds := make([]string, 0, len(ids))

								for _, id := range ids {
									// we only need to care about X25519 identities here because SSH identities are
									// considered external and are not managed by gopass. users should use ssh-keygen
									// and such to deal with them. At least we definitely don't want to remove them.
									if x, ok := id.(*age.X25519Identity); ok && x.Recipient().String() == victim {
										continue
									}
									newIds = append(newIds, fmt.Sprintf("%s", id))
								}

								return a.saveIdentities(ctx, newIds, false)
							},
						},
					},
				},
			},
		},
	}
}
