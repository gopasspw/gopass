package action

import (
	"github.com/urfave/cli/v2"
)

// GetCommands returns the cli commands exported by this module
func (s *Action) GetCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "audit",
			Usage: "Scan for weak passwords",
			Description: "" +
				"This command decrypts all secrets and checks for common flaws and (optionally) " +
				"against a list of previously leaked passwords.",
			Before: s.Initialized,
			Action: s.Audit,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "jobs",
					Aliases: []string{"j"},
					Usage:   "The number of jobs to run concurrently when auditing",
					Value:   1,
				},
			},
			Subcommands: []*cli.Command{
				{
					Name:  "hibp",
					Usage: "Detect leaked passwords",
					Description: "" +
						"This command will decrypt all secrets and check the passwords against the public " +
						"havibeenpwned.com v2 API or dumps. " +
						"To use the dumps you need to download the dumps from https://haveibeenpwned.com/passwords first. Be sure to grab the one that says '(ordered by hash)'. " +
						"This is a very expensive operation, for advanced users. " +
						"Most users should probably use the API. " +
						"If you want to use the dumps you need to use 7z to extract the dump: 7z x pwned-passwords-ordered-2.0.txt.7z.",
					Before: s.Initialized,
					Action: s.HIBP,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "force",
							Aliases: []string{"f"},
							Usage:   "Force to move the secret and overwrite existing one",
						},
						&cli.BoolFlag{
							Name:    "api",
							Aliases: []string{"a"},
							Usage:   "Use HIBP API",
						},
						&cli.StringSliceFlag{
							Name:  "dumps",
							Usage: "One or more HIBP v1/v2 dumps",
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
				"or to add a new mounted sub-store." +
				"" +
				"Needs at least one argument (git URL) to clone from. " +
				"Accepts a second argument (mount location) to clone and mount a sub-store, e.g. " +
				"'gopass clone git@example.com/store.git foo/bar'",
			Action: s.Clone,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to clone the repo to",
				},
				&cli.StringFlag{
					Name:  "crypto",
					Usage: "Select crypto backend (gpg, gpgcli, plain, xc)",
				},
				&cli.StringFlag{
					Name:  "sync",
					Usage: "Select sync backend (git, gitcli, noop)",
				},
			},
		},
		{
			Name:  "config",
			Usage: "Edit configuration",
			Description: "" +
				"This command allows for easy printing and editing of the configuration. " +
				"Without argument, the entire config is printed. " +
				"With a single argument, a single key can be printed. " +
				"With two arguments a setting specified by key can be set to value.",
			Action:       s.Config,
			BashComplete: s.ConfigComplete,
			Flags: []cli.Flag{
				&cli.StringFlag{
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
				"This also works across different sub-stores. If the source is a directory it will " +
				"automatically copy recursively. In that case, the source directory is re-created " +
				"at the destination if no trailing slash is found, otherwise the contents are " +
				"flattened (similar to rsync).",
			Before:       s.Initialized,
			Action:       s.Copy,
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Force to copy the secret and overwrite existing one",
				},
			},
		},
		{
			Name:  "delete",
			Usage: "Remove secrets",
			Description: "" +
				"This command removes secrets. It can work recursively on folders. " +
				"Recursing across stores is purposefully not supported.",
			Aliases:      []string{"remove", "rm"},
			Before:       s.Initialized,
			Action:       s.Delete,
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r"},
					Usage:   "Recursive delete files and folders",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Force to delete the secret",
				},
			},
		},
		{
			Name:  "edit",
			Usage: "Edit new or existing secrets",
			Description: "" +
				"Use this command to insert a new secret or edit an existing one using " +
				"your $EDITOR. It will attempt to create a secure temporary directory " +
				"for storing your secret while the editor is accessing it. Please make " +
				"sure your editor doesn't leak sensitive data to other locations while " +
				"editing.",
			Before:       s.Initialized,
			Action:       s.Edit,
			Aliases:      []string{"set"},
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "editor",
					Aliases: []string{"e"},
					Usage:   "Use this editor binary",
				},
				&cli.BoolFlag{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Create a new secret if none found",
				},
			},
		},
		{
			Name:  "find",
			Usage: "Search for secrets",
			Description: "" +
				"This command will first attempt a simple pattern match on the name of the " +
				"secret. If that yields no results, it will trigger a fuzzy search. " +
				"If there is an exact match it will be shown directly; if there are " +
				"multiple matches, a selection will be shown.",
			Before:       s.Initialized,
			Action:       s.FindNoFuzzy,
			Aliases:      []string{"search"},
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "clip",
					Aliases: []string{"c"},
					Usage:   "Copy the password into the clipboard",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "In the case of an exact match, display the password even if safecontent is enabled",
				},
			},
		},
		{
			Name:  "fsck",
			Usage: "Check store integrity",
			Description: "" +
				"Check the integrity of the given sub-store or all stores if none are specified. " +
				"Will automatically fix all issues found.",
			Before:       s.Initialized,
			Action:       s.Fsck,
			BashComplete: s.MountsComplete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "decrypt",
					Usage: "Decrypt during fsck",
				},
			},
		},
		{
			Name:  "generate",
			Usage: "Generate a new password",
			Description: "" +
				"Generate a new password of the specified length, optionally with no symbols. " +
				"Alternatively, a xkcd style password can be generated (https://xkcd.com/936/). " +
				"Optionally put it on the clipboard and clear clipboard after 45 seconds. " +
				"Prompt before overwriting existing password unless forced. " +
				"It will replace only the first line of an existing file with a new password.",
			Before:       s.Initialized,
			Action:       s.Generate,
			BashComplete: s.CompleteGenerate,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "clip",
					Aliases: []string{"c"},
					Usage:   "Copy the generated password to the clipboard",
					Value:   true,
				},
				&cli.BoolFlag{
					Name:    "print",
					Aliases: []string{"p"},
					Usage:   "Print the generated password to the terminal",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Force to overwrite existing password",
				},
				&cli.BoolFlag{
					Name:    "edit",
					Aliases: []string{"e"},
					Usage:   "Open secret for editing after generating a password",
				},
				&cli.BoolFlag{
					Name:    "symbols",
					Aliases: []string{"s"},
					Usage:   "Use symbols in the password",
				},
				&cli.BoolFlag{
					Name:  "strict",
					Usage: "Require strict character class rules",
				},
				&cli.BoolFlag{
					Name:    "xkcd",
					Aliases: []string{"x"},
					Usage:   "Use multiple random english words combined to a password. By default, space is used as separator and all words are lowercase",
				},
				&cli.StringFlag{
					Name:    "xkcdsep",
					Aliases: []string{"xs"},
					Usage:   "Word separator for generated xkcd style password. If no separator is specified, the words are combined without spaces/separator and the first character of words is capitalised. This flag implies -xkcd",
					Value:   "",
				},
				&cli.StringFlag{
					Name:    "xkcdlang",
					Aliases: []string{"xl"},
					Usage:   "Language to generate password from, currently de (german) and en (english, default) are supported",
					Value:   "en",
				},
			},
		},
		{
			Name:  "git",
			Usage: "Run a git command inside a password store (init, remote, push, pull)",
			Description: "" +
				"If the password store is a git repository, execute a git command " +
				"specified by git-command-args.",
			Subcommands: []*cli.Command{
				{
					Name:        "init",
					Usage:       "Init git repo",
					Description: "Create and initialize a new git repo in the store",
					Before:      s.Initialized,
					Action:      s.GitInit,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
						&cli.StringFlag{
							Name:  "sign-key",
							Usage: "GPG Key to sign commits",
						},
						&cli.StringFlag{
							Name:  "rcs",
							Usage: "Select sync backend (git, gitcli, noop)",
						},
					},
				},
				{
					Name:        "remote",
					Usage:       "Manage git remotes",
					Description: "These subcommands can be used to manage git remotes",
					Before:      s.Initialized,
					Subcommands: []*cli.Command{
						{
							Name:        "add",
							Usage:       "Add git remote",
							Description: "Add a new git remote",
							Before:      s.Initialized,
							Action:      s.GitAddRemote,
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "store",
									Usage: "Store to operate on",
								},
							},
						},
						{
							Name:        "remove",
							Usage:       "Remove git remote",
							Description: "Remove a git remote",
							Before:      s.Initialized,
							Action:      s.GitRemoveRemote,
							Flags: []cli.Flag{
								&cli.StringFlag{
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
					Before:      s.Initialized,
					Action:      s.GitPush,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
				{
					Name:        "pull",
					Usage:       "Pull from remote",
					Description: "Pull from a git remote",
					Before:      s.Initialized,
					Action:      s.GitPull,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
				{
					Name:        "status",
					Usage:       "RCS status",
					Description: "Show the RCS status",
					Before:      s.Initialized,
					Action:      s.GitStatus,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
			},
		},
		{
			Name:  "git-credential",
			Usage: `Use "!gopass git-credential $@" as git's credential.helper`,
			Description: "" +
				"This command allows you to cache your git-credentials with gopass." +
				"Activate by using `git config --global credential.helper \"!gopass git-credential $@\"`",
			Subcommands: []*cli.Command{
				{
					Name:   "get",
					Hidden: true,
					Action: s.GitCredentialGet,
					Before: s.GitCredentialBefore,
				},
				{
					Name:   "store",
					Hidden: true,
					Action: s.GitCredentialStore,
					Before: s.GitCredentialBefore,
				},
				{
					Name:   "erase",
					Hidden: true,
					Action: s.GitCredentialErase,
					Before: s.GitCredentialBefore,
				},
				{
					Name:        "configure",
					Description: "This command configures gopass as git's credential.helper",
					Action:      s.GitCredentialConfigure,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "global",
							Usage: "Install for current user",
						},
						&cli.BoolFlag{
							Name:  "local",
							Usage: "Install for current repository only",
						},
						&cli.BoolFlag{
							Name:  "system",
							Usage: "Install for all users, requires superuser rights",
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
			Before: s.Initialized,
			Action: s.Grep,
		},
		{
			Name:    "history",
			Usage:   "Show password history",
			Aliases: []string{"hist"},
			Description: "" +
				"Display the change history for a secret",
			Before:       s.Initialized,
			Action:       s.History,
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "password",
					Aliases: []string{"p"},
					Usage:   "Include passwords in output",
				},
			},
		},
		{
			Name:  "init",
			Usage: "Initialize new password store.",
			Description: "" +
				"Initialize new password storage and use gpg-id for encryption.",
			Action: s.Init,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "path",
					Aliases: []string{"p"},
					Usage:   "Set the sub-store path to operate on",
				},
				&cli.StringFlag{
					Name:    "store",
					Aliases: []string{"s"},
					Usage:   "Set the name of the sub-store",
				},
				&cli.StringFlag{
					Name:  "crypto",
					Usage: "Select crypto backend (gpg, gpgcli, plain, xc)",
				},
				&cli.StringFlag{
					Name:  "rcs",
					Usage: "Select sync backend (git, gitcli, noop)",
				},
				&cli.BoolFlag{
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
			Before:       s.Initialized,
			Action:       s.Insert,
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "echo",
					Aliases: []string{"e"},
					Usage:   "Display secret while typing",
				},
				&cli.BoolFlag{
					Name:    "multiline",
					Aliases: []string{"m"},
					Usage:   "Insert using $EDITOR",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Overwrite any existing secret and do not prompt to confirm recipients",
				},
				&cli.BoolFlag{
					Name:    "append",
					Aliases: []string{"a"},
					Usage:   "Append to any existing data",
				},
			},
		},
		{
			Name:  "list",
			Usage: "List existing secrets",
			Description: "" +
				"This command will list all existing secrets. Provide a folder prefix to list " +
				"only certain subfolders of the store.",
			Aliases:      []string{"ls"},
			Before:       s.Initialized,
			Action:       s.List,
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "limit",
					Aliases: []string{"l"},
					Usage:   "Max tree depth",
				},
				&cli.BoolFlag{
					Name:    "flat",
					Aliases: []string{"f"},
					Usage:   "Print flat list",
				},
				&cli.BoolFlag{
					Name:    "folders",
					Aliases: []string{"fo"},
					Usage:   "Print flat list of folders",
				},
				&cli.BoolFlag{
					Name:    "strip-prefix",
					Aliases: []string{"s"},
					Usage:   "Strip prefix from filtered entries",
				},
			},
		},
		{
			Name:    "move",
			Aliases: []string{"mv"},
			Usage:   "Move secrets from one location to another",
			Description: "" +
				"This command moves a secret from one path to another. This also works " +
				"across different sub-stores. If the source is a directory, the source directory " +
				"is re-created at the destination if no trailing slash is found, otherwise the " +
				"contents are flattened (similar to rsync).",
			Before:       s.Initialized,
			Action:       s.Move,
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Force to move the secret and overwrite existing one",
				},
			},
		},
		{
			Name:  "mounts",
			Usage: "Edit mounted stores",
			Description: "" +
				"This command displays all mounted password stores. It offers several " +
				"subcommands to create or remove mounts.",
			Before: s.Initialized,
			Action: s.MountsPrint,
			Subcommands: []*cli.Command{
				{
					Name:    "add",
					Aliases: []string{"mount"},
					Usage:   "Mount a password store",
					Description: "" +
						"This command allows for mounting an existing or new password store " +
						"at any path in an existing root store.",
					Before: s.Initialized,
					Action: s.MountAdd,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "init",
							Aliases: []string{"i"},
							Usage:   "Init the store with the given recipient key",
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
					Before:       s.Initialized,
					Action:       s.MountRemove,
					BashComplete: s.MountsComplete,
				},
			},
		},
		{
			Name:  "recipients",
			Usage: "Edit recipient permissions",
			Description: "" +
				"This command displays all existing recipients for all mounted stores. " +
				"The subcommands allow adding or removing recipients.",
			Before: s.Initialized,
			Action: s.RecipientsPrint,
			Subcommands: []*cli.Command{
				{
					Name:    "add",
					Aliases: []string{"authorize"},
					Usage:   "Add any number of Recipients to any store",
					Description: "" +
						"This command adds any number of recipients to any existing store. " +
						"If none are given it will display a list of usable public keys. " +
						"After adding the recipient to the list it will re-encrypt the whole " +
						"affected store to make sure the recipient has access to all existing " +
						"secrets.",
					Before: s.Initialized,
					Action: s.RecipientsAdd,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
						&cli.BoolFlag{
							Name:  "force",
							Usage: "Force adding non-existing keys",
						},
					},
				},
				{
					Name:    "remove",
					Aliases: []string{"rm", "deauthorize"},
					Usage:   "Remove any number of Recipients from any store",
					Description: "" +
						"This command removes any number of recipients from any existing store. " +
						"If no recipients are provided, it will show a list of existing recipients " +
						"to choose from. It will refuse to remove the current user's key from the " +
						"store to avoid losing access. After removing the keys it will re-encrypt " +
						"all existing secrets. Please note that the removed recipients will still " +
						"be able to decrypt old revisions of the password store and any local " +
						"copies they might have. The only way to reliably remove a recipient is to " +
						"rotate all existing secrets.",
					Before:       s.Initialized,
					Action:       s.RecipientsRemove,
					BashComplete: s.RecipientsComplete,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
						&cli.BoolFlag{
							Name:  "force",
							Usage: "Force adding non-existing keys",
						},
					},
				},
				{
					Name:  "update",
					Usage: "Recompute the saved recipient list checksums",
					Description: "" +
						"This command will recompute the saved recipient checksum" +
						"and save them to the config.",
					Before: s.Initialized,
					Action: s.RecipientsUpdate,
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
			Action: s.InitOnboarding,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "remote",
					Usage: "URL to a git remote, will attempt to join this team",
				},
				&cli.StringFlag{
					Name:  "alias",
					Usage: "Local mount point for the given remote",
				},
				&cli.BoolFlag{
					Name:  "create",
					Usage: "Create a new team (default: false, i.e. join an existing team)",
				},
				&cli.StringFlag{
					Name:  "name",
					Usage: "Firstname and Lastname for unattended GPG key generation",
				},
				&cli.StringFlag{
					Name:  "email",
					Usage: "EMail for unattended GPG key generation",
				},
				&cli.StringFlag{
					Name:  "crypto",
					Usage: "Select crypto backend (gpg, gpgcli, plain, xc)",
				},
				&cli.StringFlag{
					Name:  "rcs",
					Usage: "Select sync backend (git, gitcli, noop)",
				},
			},
		},
		{
			Name:  "show",
			Usage: "Display a secret",
			Description: "" +
				"Show an existing secret and optionally put its first line on the clipboard. " +
				"If put on the clipboard, it will be cleared after 45 seconds.",
			Before:       s.Initialized,
			Action:       s.Show,
			BashComplete: s.Complete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "clip",
					Aliases: []string{"c"},
					Usage:   "Copy the first line of the secret into the clipboard",
				},
				&cli.BoolFlag{
					Name:  "alsoclip, C",
					Usage: "Copy the first line of the secret and show everything",
				},
				&cli.BoolFlag{
					Name:  "qr",
					Usage: "Print the first line of the secret as QR Code",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Display the password even if safecontent is enabled",
				},
				&cli.BoolFlag{
					Name:    "password",
					Aliases: []string{"o"},
					Usage:   "Display only the password",
				},
				&cli.BoolFlag{
					Name:    "sync",
					Aliases: []string{"s"},
					Usage:   "Sync before attempting to display the secret",
				},
				&cli.StringFlag{
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
			Before: s.Initialized,
			Action: s.Sync,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "store",
					Aliases: []string{"s"},
					Usage:   "Select the store to sync",
				},
			},
		},
		{
			Name:  "templates",
			Usage: "Edit templates",
			Description: "" +
				"List existing templates in the password store and allow for editing " +
				"and creating them.",
			Before: s.Initialized,
			Action: s.TemplatesPrint,
			Subcommands: []*cli.Command{
				{
					Name:         "show",
					Usage:        "Show a secret template.",
					Description:  "Display an existing template",
					Aliases:      []string{"cat"},
					Before:       s.Initialized,
					Action:       s.TemplatePrint,
					BashComplete: s.TemplatesComplete,
				},
				{
					Name:         "edit",
					Usage:        "Edit secret templates.",
					Description:  "Edit an existing or new template",
					Aliases:      []string{"create", "new"},
					Before:       s.Initialized,
					Action:       s.TemplateEdit,
					BashComplete: s.TemplatesComplete,
				},
				{
					Name:         "remove",
					Aliases:      []string{"rm"},
					Usage:        "Remove secret templates.",
					Description:  "Remove an existing template",
					Before:       s.Initialized,
					Action:       s.TemplateRemove,
					BashComplete: s.TemplatesComplete,
				},
			},
		},
		{
			Name:        "unclip",
			Usage:       "Internal command to clear clipboard",
			Description: "Clear the clipboard if the content matches the checksum.",
			Action:      s.Unclip,
			Hidden:      true,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "timeout",
					Usage: "Time to wait",
				},
				&cli.BoolFlag{
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
			Action: s.Update,
			Flags: []cli.Flag{
				&cli.BoolFlag{
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
			Action: s.Version,
		},
	}
}
