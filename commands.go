package main

import (
	"context"
	"fmt"

	ap "github.com/justwatchcom/gopass/pkg/action"
	"github.com/justwatchcom/gopass/pkg/action/binary"
	"github.com/justwatchcom/gopass/pkg/action/create"
	"github.com/justwatchcom/gopass/pkg/action/xc"
	"github.com/justwatchcom/gopass/pkg/agent"
	"github.com/justwatchcom/gopass/pkg/agent/client"
	"github.com/justwatchcom/gopass/pkg/config"
	"github.com/justwatchcom/gopass/pkg/ctxutil"

	"github.com/urfave/cli"
)

func getCommands(ctx context.Context, action *ap.Action, app *cli.App) []cli.Command {
	return []cli.Command{
		{
			Name:  "agent",
			Usage: "Start gopass-agent",
			Description: "" +
				"This command start the gopass agent that will cache passphrases" +
				"so they don't have to be entered repeately.",
			Action: func(c *cli.Context) error {
				ec := make(chan error)
				go func() {
					ec <- agent.New(config.Directory()).ListenAndServe(ctx)
				}()
				select {
				case <-ctx.Done():
					return fmt.Errorf("aborted")
				case e := <-ec:
					return e
				}
			},
			Subcommands: []cli.Command{
				{
					Name:   "client",
					Usage:  "Start a simple agent test client",
					Hidden: true,
					Action: func(c *cli.Context) error {
						pw, err := client.New(config.Directory()).Passphrase(ctx, "test", "test")
						if err != nil {
							return err
						}
						fmt.Println("Passphrase:" + pw)
						return nil
					},
				},
			},
		},
		{
			Name:  "audit",
			Usage: "Scan for weak passwords",
			Description: "" +
				"This command decrypts all secrets and checks for common flaws and (optionally) " +
				"against a list of previously leaked passwords.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Audit(withGlobalFlags(ctx, c), c)
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "jobs, j",
					Usage: "The number of jobs to run concurrently when auditing",
					Value: 1,
				},
			},
			Subcommands: []cli.Command{
				{
					Name:  "hibp",
					Usage: "Detect leaked passwords",
					Description: "" +
						"This command will decrypt all secrets and check the passwords against the public " +
						"havibeenpwned.com v2 API or Dumps. " +
						"To use the dumps you need to download the dumps from https://haveibeenpwned.com/passwords first. Be sure to grap the one that says '(ordered by hash)'. " +
						"This is a very expensive operation for advanced users. " +
						"Most users should probably use the API. " +
						"If you want to use the dumps you need to use 7z to extract the dump: 7z x pwned-passwords-ordered-2.0.txt.7z.",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.HIBP(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Force to move the secret and overwrite existing one",
						},
						cli.BoolFlag{
							Name:  "api, a",
							Usage: "Use HIBP API",
						},
						cli.StringSliceFlag{
							Name:  "dumps",
							Usage: "One or more HIBP v1/v2 dumps",
						},
					},
				},
			},
		},
		{
			Name:  "binary",
			Usage: "Assist with Binary/Base64 content",
			Description: "" +
				"These commands directly convert binary files from/to base64 encoding.",
			Aliases: []string{"bin"},
			Subcommands: []cli.Command{
				{
					Name:  "cat",
					Usage: "Print content of a secret to stdout or insert from stdin",
					Description: "" +
						"This command is similar to the way cat works on the command line. " +
						"It can either be used to retrieve the decoded content of a secret " +
						"similar to 'cat file' or vice versa to encode the content from STDIN " +
						"to a secret.",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return binary.Cat(withGlobalFlags(ctx, c), c, action.Store)
					},
					BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
				},
				{
					Name:  "sum",
					Usage: "Compute the SHA256 checksum",
					Description: "" +
						"This command decodes an Base64 encoded secret and computes the SHA256 checksum " +
						"over the decoded data. This is useful to verify the integrity of an " +
						"inserted secret.",
					Aliases: []string{"sha", "sha256"},
					Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return binary.Sum(withGlobalFlags(ctx, c), c, action.Store)
					},
					BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
				},
				{
					Name:  "copy",
					Usage: "Copy files from or to the password store",
					Description: "" +
						"This command either reads a file from the filesystem and writes the " +
						"encoded and encrypted version in the store or it decrypts and decodes " +
						"a secret and write the result to a file. Either source or destination " +
						"must be a file and the other one a secret. If you want the source to " +
						"be securely removed after copying use 'gopass binary move'",
					Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Aliases: []string{"cp"},
					Action: func(c *cli.Context) error {
						return binary.Copy(withGlobalFlags(ctx, c), c, action.Store)
					},
					BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Force to move the secret and overwrite existing one",
						},
					},
				},
				{
					Name:  "move",
					Usage: "Move files from or to the password store",
					Description: "" +
						"This command either reads a file from the filesystem and writes the " +
						"encoded and encrypted version in the store or it decrypts and decodes " +
						"a secret and write the result to a file. Either source or destination " +
						"must be a file and the other one a secret. The source will be wiped " +
						"from disk or from the store after it has been copied successfully " +
						"and validated. If you don't want the source to be removed use " +
						"'gopass binary copy'",
					Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Aliases: []string{"mv"},
					Action: func(c *cli.Context) error {
						return binary.Move(withGlobalFlags(ctx, c), c, action.Store)
					},
					BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Force to move the secret and overwrite existing one",
						},
					},
				},
			},
		},
		{
			Name:  "clone",
			Usage: "Clone a store from git",
			Description: "" +
				"This command clones an existing password store from a git remote to " +
				"a local password store. Can be either used to initialize a new root store " +
				"or to add a new mounted sub store." +
				"" +
				"Needs at least one argument (git URL) to clone from. " +
				"Accepts as second argument (mount location) to clone and mount a sub store, e.g. " +
				"gopass clone git@example.com/store.git foo/bar",
			Action: func(c *cli.Context) error {
				return action.Clone(withGlobalFlags(ctx, c), c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path",
					Usage: "Path to clone the repo to",
				},
				cli.StringFlag{
					Name:  "crypto",
					Usage: "Select crypto backend (gpg, gpgcli, plain, xc)",
				},
				cli.StringFlag{
					Name:  "sync",
					Usage: "Select sync backend (git, gitcli, gogit, noop)",
				},
			},
		},
		{
			Name:  "completion",
			Usage: "Bash and ZSH completion",
			Description: "" +
				"Source the output of this command with bash or zsh to get auto completion",
			Subcommands: []cli.Command{{
				Name:   "bash",
				Usage:  "Source for auto completion in bash",
				Action: action.CompletionBash,
			}, {
				Name:  "zsh",
				Usage: "Source for auto completion in zsh",
				Action: func(c *cli.Context) error {
					return action.CompletionZSH(c, app)
				},
			}, {
				Name:  "fish",
				Usage: "Source for auto completion in fish",
				Action: func(c *cli.Context) error {
					return action.CompletionFish(c, app)
				},
			}, {
				Name:  "openbsdksh",
				Usage: "Source for auto completion in OpenBSD's ksh",
				Action: func(c *cli.Context) error {
					return action.CompletionOpenBSDKsh(c, app)
				},
			}},
		},
		{
			Name:  "config",
			Usage: "Edit configuration",
			Description: "" +
				"This command allows for easy editing of the configuration",
			Action: func(c *cli.Context) error {
				return action.Config(withGlobalFlags(ctx, c), c)
			},
			BashComplete: action.ConfigComplete,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "store",
					Usage: "Set value to substore config",
				},
			},
		},
		{
			Name:    "copy",
			Aliases: []string{"cp"},
			Usage:   "Copy secrets from one location to another",
			Description: "" +
				"This command copies an existing secret in the store to another location. " +
				"It will also handle copying secrets to different sub stores. " +
				"If the destination is directory it will automatically copy recursively.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Copy(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to copy the secret and overwrite existing one",
				},
			},
		},
		{
			Name:    "create",
			Aliases: []string{"new"},
			Usage:   "Easy creation of new secrets",
			Description: "" +
				"This command starts a wizard to aid in creation of new secrets.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return create.Create(withGlobalFlags(ctx, c), c, action.Store)
			},
		},
		{
			Name:  "delete",
			Usage: "Remove secrets",
			Description: "" +
				"This command removes secrets. It can work recursively on folders. " +
				"Recursing across stores is purposefully not supported.",
			Aliases: []string{"remove", "rm"},
			Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Delete(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "recursive, r",
					Usage: "f",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to delete the secret",
				},
			},
		},
		{
			Name:  "edit",
			Usage: "Edit new or existing secret",
			Description: "" +
				"Use this command to insert a new secret or edit an existing one using " +
				"your $EDITOR. It will attempt to create a secure temporary directory " +
				"for storing your secret while the editor is accessing it. Please make " +
				"sure your editor doesn't leak sensitive data to other locations while " +
				"editing.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Edit(withGlobalFlags(ctx, c), c)
			},
			Aliases:      []string{"set"},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "editor, e",
					Usage: "Use this editor binary",
				},
			},
		},
		{
			Name:  "find",
			Usage: "Search for secrets",
			Description: "" +
				"This command will first attempt a simple pattern match on the name of the " +
				"secret. If that yields no results it will trigger a fuzzy search. " +
				"If there is an exact match it will be shown diretly, if there are " +
				"multiple matches a selection will be shown.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Find(withGlobalFlags(ctxutil.WithFuzzySearch(ctx, false), c), c)
			},
			Aliases:      []string{"search"},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the password into the clipboard",
				},
			},
		},
		{
			Name:  "fsck",
			Usage: "Check store integrity",
			Description: "" +
				"Check the integrity of the given sub store or all stores if none are specified. " +
				"Will automatically fix all issues found.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Fsck(withGlobalFlags(ctx, c), c)
			},
			BashComplete: action.MountsComplete,
		},
		{
			Name:  "generate",
			Usage: "Generate a new password",
			Description: "" +
				"Generate a new password of the specified length with optionally no symbols. " +
				"Alternatively, a xkcd style password can be generated (https://xkcd.com/936/). " +
				"Optionally put it on the clipboard and clear board after 45 seconds. " +
				"Prompt before overwriting existing password unless forced. " +
				"It will replace only the first line of an existing file with a new password.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Generate(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the generated password to the clipboard",
				},
				cli.BoolFlag{
					Name:  "print, p",
					Usage: "Print the generated password to the terminal",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to overwrite existing password",
				},
				cli.BoolFlag{
					Name:  "edit, e",
					Usage: "Open secret for editing after generating a password",
				},
				cli.BoolFlag{
					Name:  "symbols, s",
					Usage: "Use symbols in the password",
				},
				cli.BoolFlag{
					Name:  "xkcd, x",
					Usage: "Use multiple random english words combined to a password. If no separator is specified, the words are combined without spaces/separator and the first character of words is capitalised",
				},
				cli.StringFlag{
					Name:  "xkcdsep, xs",
					Usage: "Word separator for generated xkcd style password. Implies -xkcd",
					Value: "",
				},
				cli.StringFlag{
					Name:  "xkcdlang, xl",
					Usage: "Language to generate password from, currently de (german) and en (english, default) are supported",
					Value: "en",
				},
			},
		},
		{
			Name:        "jsonapi",
			Usage:       "Run gopass as jsonapi e.g. for browser plugins",
			Description: "Setup and run gopass as native messaging hosts, e.g. for browser plugins.",
			Hidden:      true,
			Subcommands: []cli.Command{
				{
					Name:        "listen",
					Usage:       "Listen and respond to messages via stdin/stdout",
					Description: "Gopass is started in listen mode from browser plugins using a wrapper specified in native messaging host manifests",
					Action: func(c *cli.Context) error {
						return action.JSONAPI(withGlobalFlags(ctx, c), c)
					},
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
				},
				{
					Name:        "configure",
					Usage:       "Setup gopass native messaging manifest for selected browser",
					Description: "To access gopass from browser plugins, a native app manifest must be installed at the correct location",
					Action: func(c *cli.Context) error {
						return action.SetupNativeMessaging(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "browser",
							Usage: "One of 'chrome' and 'firefox'",
						},
						cli.StringFlag{
							Name:  "path",
							Usage: "Path to install 'gopass_wrapper.sh' to",
						},
						cli.StringFlag{
							Name:  "manifest-path",
							Usage: "Path to install 'com.justwatch.gopass.json' to",
						},
						cli.BoolFlag{
							Name:  "global",
							Usage: "Install for all users, requires superuser rights",
						},
						cli.StringFlag{
							Name:  "libpath",
							Usage: "Library path for global installation on linux. Default is /usr/lib",
						},
						cli.StringFlag{
							Name:  "gopass-path",
							Usage: "Path to gopass binary. Default is auto detected",
						},
						cli.BoolTFlag{
							Name:  "print",
							Usage: "Print installation summary before creating any files",
						},
					},
				},
			},
		},
		{
			Name:    "otp",
			Usage:   "Generate time or hmac based tokens",
			Aliases: []string{"totp", "hotp"},
			Description: "" +
				"Tries to parse an OTP URL (otpauth://). " +
				"URL can be TOTP or HOTP.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.OTP(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the time based token into the clipboard",
				},
				cli.StringFlag{
					Name:  "qr, q",
					Usage: "Write QR code to `FILE`",
				},
			},
		},
		{
			Name:  "git",
			Usage: "Run any git command inside a password store",
			Description: "" +
				"If the password store is a git repository, execute a git command " +
				"specified by git-command-args.",
			Subcommands: []cli.Command{
				{
					Name:        "init",
					Usage:       "Init git repo",
					Description: "Create and initialize a new git repo in the store",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.GitInit(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
						cli.StringFlag{
							Name:  "sign-key",
							Usage: "GPG Key to sign commits",
						},
						cli.StringFlag{
							Name:  "rcs",
							Usage: "Select sync backend (git, gitcli, gogit, noop)",
						},
					},
				},
				{
					Name:        "remote",
					Usage:       "Mangage git remotes",
					Description: "These subcommands can be used to manage git remotes",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Subcommands: []cli.Command{
						{
							Name:        "add",
							Usage:       "Add git remote",
							Description: "Add a new git remote",
							Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
							Action: func(c *cli.Context) error {
								return action.GitAddRemote(withGlobalFlags(ctx, c), c)
							},
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "store",
									Usage: "Store to operate on",
								},
							},
						},
						{
							Name:        "remove",
							Usage:       "Remove git remote",
							Description: "Remove a git remote",
							Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
							Action: func(c *cli.Context) error {
								return action.GitRemoveRemote(withGlobalFlags(ctx, c), c)
							},
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "store",
									Usage: "Store to operate on",
								},
							},
						},
					},
				},
				{
					Name:        "push",
					Usage:       "Push to remote",
					Description: "Push to a git remote",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.GitPush(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
				{
					Name:        "pull",
					Usage:       "Pull from remote",
					Description: "Pull from a git remote",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.GitPull(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
			},
		},
		{
			Name:  "grep",
			Usage: "Search for secrets files containing search-string when decrypted.",
			Description: "" +
				"This command decrypts all secrets and performs a pattern matching on the " +
				"content.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Grep(withGlobalFlags(ctx, c), c)
			},
		},
		{
			Name:    "history",
			Usage:   "Show password history",
			Aliases: []string{"hist"},
			Description: "" +
				"Display the change history for a secret",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.History(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "password, p",
					Usage: "Include passwords in output",
				},
			},
		},
		{
			Name:  "init",
			Usage: "Initialize new password store.",
			Description: "" +
				"Initialize new password storage and use gpg-id for encryption.",
			Action: func(c *cli.Context) error {
				return action.Init(withGlobalFlags(ctx, c), c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path, p",
					Usage: "Set the sub store path to operate on",
				},
				cli.StringFlag{
					Name:  "store, s",
					Usage: "Set the name of the sub store",
				},
				cli.StringFlag{
					Name:  "crypto",
					Usage: "Select crypto backend (gpg, gpgcli, plain, xc)",
				},
				cli.StringFlag{
					Name:  "rcs",
					Usage: "Select sync backend (git, gitcli, gogit, noop)",
				},
				cli.BoolFlag{
					Name:   "nogit",
					Usage:  "(DEPRECATED): Select noop RCS backend. Use '--rcs noop' instead",
					Hidden: true,
				},
			},
		},
		{
			Name:  "insert",
			Usage: "Insert a new secret",
			Description: "" +
				"Insert a new secret. Optionally, echo the secret back to the console during entry. " +
				"Or, optionally, the entry may be multiline. " +
				"Prompt before overwriting existing secret unless forced.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Insert(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "echo, e",
					Usage: "Display secret while typing",
				},
				cli.BoolFlag{
					Name:  "multiline, m",
					Usage: "Insert using $EDITOR",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Overwrite any existing secret and do not prompt to confirm recipients",
				},
			},
		},
		{
			Name:  "list",
			Usage: "List existing secrets",
			Description: "" +
				"This command will list all existing secrets. Provide a folder prefix to list " +
				"only certain subfolders of the store.",
			Aliases: []string{"ls"},
			Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.List(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "limit, l",
					Usage: "Max tree depth",
				},
				cli.BoolFlag{
					Name:  "flat, f",
					Usage: "Print flat list",
				},
				cli.BoolFlag{
					Name:  "strip-prefix, s",
					Usage: "Strip prefix from filtered entries",
				},
			},
		},
		{
			Name:    "move",
			Aliases: []string{"mv"},
			Usage:   "Move secrets from one location to another",
			Description: "" +
				"This command moves a secret from one path to another. This works even " +
				"across different sub stores.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Move(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to move the secret and overwrite existing one",
				},
			},
		},
		{
			Name:  "mounts",
			Usage: "Edit mounted stores",
			Description: "" +
				"This command displays all mounted password stores. It offers several " +
				"subcommands to create or remove mounts.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.MountsPrint(withGlobalFlags(ctx, c), c)
			},
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"mount"},
					Usage:   "Mount a password store",
					Description: "" +
						"This command allows for mounting an existing or new password store " +
						"at any path in an existing root store.",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.MountAdd(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "init, i",
							Usage: "Init the store with the given recipient key",
						},
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"rm", "unmount", "umount"},
					Usage:   "Umount an mounted password store",
					Description: "" +
						"This command allows to unmount an mounted password store. This will " +
						"only updated the configuration and not delete the password store.",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.MountRemove(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.MountsComplete,
				},
			},
		},
		{
			Name:  "recipients",
			Usage: "Edit recipient permissions",
			Description: "" +
				"This command displays all existing recipients for all mounted stores. " +
				"The subcommands allow adding or removing recipients.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.RecipientsPrint(withGlobalFlags(ctx, c), c)
			},
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"authorize"},
					Usage:   "Add any number of Recipients to any store",
					Description: "" +
						"This command adds any number of recipients to any existing store. " +
						"If none are given it will display a list of useable public keys. " +
						"After adding the recipient to the list it will reencrypt the whole " +
						"affected store to make sure the recipient has access to any existing " +
						"secret.",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.RecipientsAdd(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"rm", "deauthorize"},
					Usage:   "Remove any number of Recipients from any store",
					Description: "" +
						"This command removes any number of recipients from any existing store. " +
						"If no recipients are provided it will show a list of existing recipients " +
						"to choose from. It will refuse to remove the current users key from the " +
						"store to avoid loosing access. After removing the keys it will re-encrypt " +
						"all existing secrets. Please note that the removed recipients will still " +
						"be able to decrypt old revisions of the password store and any local " +
						"copies they might have. The only way to reliably remove a recipients is to " +
						"rotate all existing secrets.",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.RecipientsRemove(withGlobalFlags(ctx, c), c)
					},
					BashComplete: func(c *cli.Context) {
						action.RecipientsComplete(ctx, c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
			},
		},
		{
			Name:  "setup",
			Usage: "Initialize a new password store",
			Description: "" +
				"This command is automatically invoked if gopass is started without any " +
				"existing password store. This command exists so users can be provided with " +
				"simple one-command setup instructions.",
			Hidden: true,
			Action: func(c *cli.Context) error {
				return action.InitOnboarding(withGlobalFlags(ctx, c), c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "remote",
					Usage: "URL to a git remote, will attempt to join this team",
				},
				cli.StringFlag{
					Name:  "alias",
					Usage: "Local mount point for the given remote",
				},
				cli.BoolFlag{
					Name:  "create",
					Usage: "Create a new team (default: false, i.e. join an existing team)",
				},
				cli.StringFlag{
					Name:  "name",
					Usage: "Firstname and Lastname for unattended GPG key generation",
				},
				cli.StringFlag{
					Name:  "email",
					Usage: "EMail for unattended GPG key generation",
				},
				cli.StringFlag{
					Name:  "crypto",
					Usage: "Select crypto backend (gpg, gpgcli, plain, xc)",
				},
				cli.StringFlag{
					Name:  "rcs",
					Usage: "Select sync backend (git, gitcli, gogit, noop)",
				},
			},
		},
		{
			Name:  "show",
			Usage: "Display a secret",
			Description: "" +
				"Show an existing secret and optionally put its first line on the clipboard. " +
				"If put on the clipboard, it will be cleared after 45 seconds.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Show(withGlobalFlags(ctx, c), c)
			},
			BashComplete: func(c *cli.Context) { action.Complete(ctx, c) },
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the first line of the secret into the clipboard",
				},
				cli.BoolFlag{
					Name:  "qr",
					Usage: "Print the first line of the secret as QR Code",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Display the password even if safecontent is enabled",
				},
				cli.BoolFlag{
					Name:  "password, o",
					Usage: "Display only the password",
				},
				cli.BoolFlag{
					Name:  "sync, s",
					Usage: "Sync before attempting to display the secret",
				},
				cli.StringFlag{
					Name:  "revision",
					Usage: "Show a past revision",
				},
			},
		},
		{
			Name:  "sync",
			Usage: "Sync all local stores with their remotes",
			Description: "" +
				"Sync all local stores with their git remotes, if any, and check " +
				"any possibly affected gpg keys.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Sync(withGlobalFlags(ctx, c), c)

			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "store, s",
					Usage: "Select the store to sync",
				},
			},
		},
		{
			Name:  "templates",
			Usage: "Edit templates",
			Description: "" +
				"List existing templates in the password store and allow for editing " +
				"and creating them.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.TemplatesPrint(withGlobalFlags(ctx, c), c)
			},
			Subcommands: []cli.Command{
				{
					Name:        "show",
					Usage:       "Show a secret template.",
					Description: "Display an existing template",
					Aliases:     []string{"cat"},
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.TemplatePrint(withGlobalFlags(ctx, c), c)
					},
					BashComplete: func(c *cli.Context) { action.TemplatesComplete(ctx, c) },
				},
				{
					Name:        "edit",
					Usage:       "Edit secret templates.",
					Description: "Edit an existing or new template",
					Aliases:     []string{"create", "new"},
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.TemplateEdit(withGlobalFlags(ctx, c), c)
					},
					BashComplete: func(c *cli.Context) { action.TemplatesComplete(ctx, c) },
				},
				{
					Name:        "remove",
					Aliases:     []string{"rm"},
					Usage:       "Remove secret templates.",
					Description: "Remove an existing template",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.TemplateRemove(withGlobalFlags(ctx, c), c)
					},
					BashComplete: func(c *cli.Context) { action.TemplatesComplete(ctx, c) },
				},
			},
		},
		{
			Name:        "unclip",
			Usage:       "Internal command to clear clipboard",
			Description: "Clear the clipboard if the content matches the checksum.",
			Action: func(c *cli.Context) error {
				return action.Unclip(withGlobalFlags(ctx, c), c)
			},
			Hidden: true,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "timeout",
					Usage: "Time to wait",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "Clear clipboard even if checksum mismatches",
				},
			},
		},
		{
			Name:  "update",
			Usage: "Check for updates",
			Description: "" +
				"This command checks for gopass updates at GitHub and automatically " +
				"downloads and installs any missing update.",
			Action: func(c *cli.Context) error {
				return action.Update(withGlobalFlags(ctx, c), c)
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "pre",
					Usage: "Update to prereleases",
				},
			},
		},
		{
			Name:  "version",
			Usage: "Display version",
			Description: "" +
				"This command displays version and build time information " +
				"along with version information of important external commands. " +
				"Please provide the output when reporting issues.",
			Action: func(c *cli.Context) error {
				return action.Version(withGlobalFlags(ctx, c), c)
			},
		},
		{
			Name:  "xc",
			Usage: "Experimental Crypto",
			Description: "" +
				"These subcommands are used to control and test the experimental crypto" +
				"implementation.",
			Subcommands: []cli.Command{
				{
					Name: "list-private-keys",
					Action: func(c *cli.Context) error {
						return xc.ListPrivateKeys(withGlobalFlags(ctx, c), c)
					},
				},
				{
					Name: "list-public-keys",
					Action: func(c *cli.Context) error {
						return xc.ListPublicKeys(withGlobalFlags(ctx, c), c)
					},
				},
				{
					Name: "generate",
					Action: func(c *cli.Context) error {
						return xc.GenerateKeypair(withGlobalFlags(ctx, c), c)
					},
				},
				{
					Name: "export",
					Action: func(c *cli.Context) error {
						return xc.ExportPublicKey(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "id",
						},
						cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name: "import",
					Action: func(c *cli.Context) error {
						return xc.ImportPublicKey(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "id",
						},
						cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name: "export-private-key",
					Action: func(c *cli.Context) error {
						return xc.ExportPrivateKey(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "id",
						},
						cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name: "import-private-key",
					Action: func(c *cli.Context) error {
						return xc.ImportPrivateKey(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "id",
						},
						cli.StringFlag{
							Name: "file",
						},
					},
				},
				{
					Name: "remove",
					Action: func(c *cli.Context) error {
						return xc.RemoveKey(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "id",
						},
					},
				},
				{
					Name: "encrypt",
					Action: func(c *cli.Context) error {
						return xc.EncryptFile(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "file",
						},
						cli.StringSliceFlag{
							Name: "recipients",
						},
						cli.BoolFlag{
							Name: "stream",
						},
					},
				},
				{
					Name: "decrypt",
					Action: func(c *cli.Context) error {
						return xc.DecryptFile(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "file",
						},
						cli.BoolFlag{
							Name: "stream",
						},
					},
				},
			},
		},
	}
}
