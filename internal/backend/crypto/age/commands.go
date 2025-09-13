package age

import (
	"fmt"
	"strings"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

//nolint:cyclop
func (l loader) Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   name,
			Hidden: false,
			Usage:  "age commands",
			Description: "" +
				"Built-in commands for the age backend.\n" +
				"These allow limited interactions with the gopass specific age identities.\n " +
				"Added identities are automatically added as recipient to your secrets when encrypting, but not to" +
				"your recipients, make sure to keep your recipients and identities in sync as you want to.\n" +
				"All age identities, including plugin ones should be supported. We also still support github" +
				"identities despite them being deprecated by age, we do so by falling back to the ssh identities" +
				"for these and keeping a local cache of ssh keys for a given github identity.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "age-ssh-key-path",
					Usage:   "Custom path to SSH key or directory for age backend",
					EnvVars: []string{"GOPASS_SSH_DIR"},
				},
			},
			Subcommands: []*cli.Command{
				{
					Name:   "agent",
					Usage:  "Start the age agent",
					Action: l.agent,
				},
				{
					Name:  "identities",
					Usage: "List age identities used for decryption and encryption",
					Description: "" +
						"List identities",
					Action: func(c *cli.Context) error {
						ctx := ctxutil.WithGlobalFlags(c)
						sshKeyPath := config.String(ctx, "age.ssh-key-path")
						if sv := c.String("age-ssh-key-path"); sv != "" {
							sshKeyPath = sv
						}
						a, err := New(ctx, sshKeyPath)
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

						for _, id := range recipientsToString(ids) {
							out.Print(ctx, out.Secret(id))
						}

						return nil
					},
					Subcommands: []*cli.Command{
						{
							Name:  "add",
							Usage: "Add an existing age identity",
							Description: "" +
								"Add an existing age identity, interactively",
							Action: func(c *cli.Context) error {
								ctx := ctxutil.WithGlobalFlags(c)
								sshKeyPath := config.String(ctx, "age.ssh-key-path")
								if sv := c.String("age-ssh-key-path"); sv != "" {
									sshKeyPath = sv
								}
								a, err := New(ctx, sshKeyPath)
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to create age backend")
								}

								idS, recEncm := c.Args().Get(0), c.Args().Get(1)

								if len(idS) < 1 {
									idS, err = termio.AskForPassword(ctx, "the age identity starting in AGE-", false)
									if err != nil {
										return exit.Error(exit.Unknown, err, "failed to read age identity")
									}
								}
								if len(recEncm) < 1 && !strings.HasPrefix(idS, "AGE-SECRET-KEY-1") {
									recEncm, err = termio.AskForString(ctx, "Provide the corresponding age recipient", "")
									if err != nil || recEncm == "" {
										return exit.Error(exit.Unknown, err, "failed to read corresponding age recipient")
									}
									if strings.HasPrefix(recEncm, "AGE-") {
										out.Warning(ctx, "You have provided an identity as a recipient, recipients should start in 'age1', this might not be properly supported and might leak secret data in our identity recipient cache")
									}
								}

								id, err := parseIdentity(idS + "|" + recEncm)
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to parse age identity")
								}

								err = a.addIdentity(ctx, id)
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to save age identity")
								}

								rec := IdentityToRecipient(id)
								out.Noticef(ctx, "New age identities are not automatically added to your recipient list, consider adding it using 'gopass recipients add %s'", rec)
								out.Warning(ctx, "If you do not add this recipient to the recipient list, make sure to re-encrypt using 'gopass fsck --decrypt' to properly support this identity")

								return nil
							},
						},
						{
							Name:  "keygen",
							Usage: "Generate a new age identity",
							Description: "" +
								"Generate a new age identity",
							Action: func(c *cli.Context) error {
								ctx := ctxutil.WithGlobalFlags(c)
								sshKeyPath := config.String(ctx, "age.ssh-key-path")
								if sv := c.String("age-ssh-key-path"); sv != "" {
									sshKeyPath = sv
								}
								a, err := New(ctx, sshKeyPath)
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to create age backend")
								}

								rec, err := a.GenerateIdentity(ctx, "", "", "")
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to generate age identity")
								}

								out.Printf(ctx, "New age identity created: %s", rec)
								out.Notice(ctx, "New age identities are not automatically added to your recipient list, consider adding it using 'gopass recipients add age1...'")
								out.Warning(ctx, "If you do not add this recipient to the recipient list, make sure to re-encrypt using 'gopass fsck --decrypt' to properly support this identity")

								return nil
							},
						},
						{
							Name:    "remove",
							Aliases: []string{"rm"},
							Usage:   "Remove an identity",
							Description: "" +
								"Remove all identity matching the argument",
							Action: func(c *cli.Context) error {
								ctx := ctxutil.WithGlobalFlags(c)
								sshKeyPath := config.String(ctx, "age.ssh-key-path")
								if sv := c.String("age-ssh-key-path"); sv != "" {
									sshKeyPath = sv
								}
								a, err := New(ctx, sshKeyPath)
								if err != nil {
									return exit.Error(exit.Unknown, err, "failed to create age backend")
								}
								victim := c.Args().First()
								if len(victim) == 0 {
									return exit.Error(exit.Usage, err, "missing argument to remove")
								}

								ids, _ := a.Identities(ctx)
								newIds := make([]string, 0, len(ids))

								debug.Log("ranging over %d age identities", len(ids))
								for _, id := range ids {
									// we only need to care about X25519 and plugin/wrapped identities here because
									// SSH identities are considered external and are not managed by gopass.
									// Users should use ssh-keygen and such to deal with them.
									// At least we definitely don't want to remove them.
									switch x := id.(type) {
									case *age.X25519Identity:
										if x.Recipient().String() == victim {
											debug.Log("will remove X25519Identity %s", x.Recipient())

											continue
										}
									case *wrappedIdentity:
										skip := false
										// to avoid fuzzy matching, let's match on entire parts
										for _, part := range strings.Split(x.String(), "|") {
											if part == victim {
												skip = true
											}
										}
										if skip {
											debug.Log("will remove Plugin Identity %s", x)

											continue
										}
									}

									newIds = append(newIds, fmt.Sprintf("%s", id))
								}
								if len(newIds) != len(ids) {
									out.Warning(ctx, "Make sure to run 'gopass fsck --decrypt' to re-encrypt your secrets without including that identity if it's not in your recipient list.")
								} else {
									out.Notice(ctx, "no matching identity found in list")
								}

								// we invalidate our recipient id cache when we remove an identity, if there's one
								if err := a.recpCache.Remove(idRecpCacheKey); err != nil {
									debug.Log("error invalidating age id recipient cache: %s", err)
								}

								return a.saveIdentities(ctx, newIds, false)
							},
						},
					},
				},
				{
					Name:   "lock",
					Usage:  "Lock the age agent",
					Action: l.lock,
				},
			},
		},
	}
}

func (l loader) agent(c *cli.Context) error {
	out.Printf(c.Context, "Starting age agent ...")

	ag, err := agent.New()
	if err != nil {
		return err
	}

	return ag.Run(c.Context)
}

func (l loader) lock(c *cli.Context) error {
	client := agent.NewClient()
	if err := client.Lock(); err != nil {
		return exit.Error(exit.Unknown, err, "failed to lock agent: %s", err)
	}

	return nil
}
