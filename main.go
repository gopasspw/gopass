package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/action"
	colorable "github.com/mattn/go-colorable"
	"github.com/urfave/cli"
)

const (
	name = "gopass"
)

var (
	// Version is the released version of gopass
	version string
	// BuildTime is the time the binary was built
	date string
	// Commit is the git hash the binary was built from
	commit string
)

type errorWriter struct {
	out io.Writer
}

func (e errorWriter) Write(p []byte) (int, error) {
	return e.out.Write([]byte("\n" + color.RedString("Error: "+string(p))))
}

func main() {
	cli.ErrWriter = errorWriter{
		out: colorable.NewColorableStderr(),
	}

	sv, err := semver.Parse(version)
	if err != nil {
		sv = semver.Version{
			Build: []string{"HEAD"},
		}
	}

	cli.VersionPrinter = func(c *cli.Context) {
		buildtime := ""
		if bt, err := time.Parse("2006-01-02T15:04:05-0700", date); err == nil {
			buildtime = bt.Format("2006-01-02 15:04:05")
		}
		buildInfo := ""
		if commit != "" {
			buildInfo = commit
		}
		if buildtime != "" {
			if buildInfo != "" {
				buildInfo += " "
			}
			buildInfo += buildtime
		}
		if buildInfo != "" {
			buildInfo = "(" + buildInfo + ") "
		}
		fmt.Printf("%s %s %s%s %s %s\n",
			name,
			sv.String(),
			buildInfo,
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		)
	}

	action := action.New(sv)
	app := cli.NewApp()

	app.Name = name
	app.Version = sv.String()
	app.Usage = "The standard unix password manager - rewritten in Go"
	app.EnableBashCompletion = true
	app.BashComplete = func(c *cli.Context) {
		cli.DefaultAppComplete(c)
		action.Complete(c)
	}

	app.Action = func(c *cli.Context) error {
		if err := action.Initialized(c); err != nil {
			return err
		}

		if c.Args().Present() {
			return action.Show(c)
		}
		return action.List(c)
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "clip, c",
			Usage: "Copy the first line of the secret into the clipboard",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "audit",
			Usage:       "Audit passwords for common flaws",
			Description: "To check passwords for common flaws (e.g. too short or from a dictionary)",
			Before:      action.Initialized,
			Action:      action.Audit,
		},
		{
			Name:    "binary",
			Usage:   "Work with binary blobs",
			Aliases: []string{"bin"},
			Subcommands: []cli.Command{
				{
					Name:         "cat",
					Usage:        "Print content of a secret to stdout or insert from stdin",
					Before:       action.Initialized,
					Action:       action.BinaryCat,
					BashComplete: action.Complete,
				},
				{
					Name:         "sum",
					Usage:        "Compute the SHA256 sum of a decoded secret",
					Aliases:      []string{"sha", "sha256"},
					Before:       action.Initialized,
					Action:       action.BinarySum,
					BashComplete: action.Complete,
				},
				{
					Name:         "copy",
					Usage:        "Copy files from or to the password store",
					Before:       action.Initialized,
					Aliases:      []string{"cp"},
					Action:       action.BinaryCopy,
					BashComplete: action.Complete,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Force to move the secret and overwrite existing one",
						},
					},
				},
				{
					Name:         "move",
					Usage:        "Move files from or to the password store",
					Before:       action.Initialized,
					Aliases:      []string{"mv"},
					Action:       action.BinaryMove,
					BashComplete: action.Complete,
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
			Name:        "clone",
			Usage:       "Clone a new store",
			Description: "To clone a remote repo",
			Action:      action.Clone,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "path",
					Usage: "Path to clone the repo to",
				},
			},
		},
		{
			Name:  "completion",
			Usage: "Source the output with bash or zsh to get auto completion",
			Subcommands: []cli.Command{{
				Name:   "bash",
				Usage:  "Source for auto completion in bash",
				Action: action.CompletionBash,
			}, {
				Name:   "zsh",
				Usage:  "Source for auto completion in zsh",
				Action: action.CompletionZSH,
			}},
		},
		{
			Name:        "config",
			Usage:       "Edit configuration",
			Description: "To manipulate the gopass configuration",
			Action:      action.Config,
		},
		{
			Name:         "copy",
			Aliases:      []string{"cp"},
			Usage:        "Copies old-path to new-path, optionally forcefully, selectively reencrypting.",
			Before:       action.Initialized,
			Action:       action.Copy,
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to copy the secret and overwrite existing one",
				},
			},
		},
		{
			Name:         "delete",
			Usage:        "Remove existing secret or directory, optionally forcefully.",
			Aliases:      []string{"remove", "rm"},
			Before:       action.Initialized,
			Action:       action.Delete,
			BashComplete: action.Complete,
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
			Name:         "edit",
			Usage:        "Insert a new secret or edit an existing secret using $EDITOR.",
			Before:       action.Initialized,
			Action:       action.Edit,
			Aliases:      []string{"set"},
			BashComplete: action.Complete,
		},
		{
			Name:         "find",
			Usage:        "List secrets that match the search term.",
			Before:       action.Initialized,
			Action:       action.Find,
			Aliases:      []string{"search"},
			BashComplete: action.Complete,
		},
		{
			Name:        "fsck",
			Usage:       "Check store integrity",
			Description: "Check integrity of all stores",
			Before:      action.Initialized,
			Action:      action.Fsck,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "check, c",
					Usage: "Only report",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Auto-correct any errors, do not ask",
				},
			},
		},
		{
			Name:  "generate",
			Usage: "Generate a new password of the specified length with optionally no symbols.",
			Description: "" +
				"Generate a new password of the specified length with optionally no symbols. " +
				"Optionally put it on the clipboard and clear board after 45 seconds. " +
				"Prompt before overwriting existing password unless forced. " +
				"Optionally replace only the first line of an existing file with a new password.",
			Before:       action.Initialized,
			Action:       action.Generate,
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the password into the clipboard",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to overwrite existing password",
				},
				cli.BoolFlag{
					Name:  "no-symbols, n",
					Usage: "Don't use symbols in the password",
				},
			},
		},
		{
			Name:         "totp",
			Usage:        "Generate time based token from stored secret",
			Description:  "Tries to parse the saved string as a time-based one-time password secret and generate a token based on the current time",
			Before:       action.Initialized,
			Action:       action.TOTP,
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the time based token into the clipboard",
				},
			},
		},
		{
			Name:        "git",
			Usage:       "Do git things",
			Description: "If the password store is a git repository, execute a git command specified by git-command-args.",
			Before:      action.Initialized,
			Action:      action.Git,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "store, s",
					Usage: "Store to operate on",
				},
				cli.BoolFlag{
					Name:  "no-recurse, n",
					Usage: "Do not recurse to mounted sub-stores",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Print errors but continue",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:        "init",
					Usage:       "Init git repo",
					Description: "Create and initialize a new git repo in the store",
					Before:      action.Initialized,
					Action:      action.GitInit,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
						cli.StringFlag{
							Name:  "sign-key",
							Usage: "GPG Key to sign commits",
						},
					},
				},
			},
		},
		{
			Name:   "grep",
			Usage:  "Search for secrets files containing search-string when decrypted.",
			Before: action.Initialized,
			Action: action.Grep,
		},
		{
			Name:  "init",
			Usage: "Initialize new password storage and use gpg-id for encryption.",
			Description: "" +
				"Initialize new password storage and use gpg-id for encryption.",
			Action: action.Init,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "store, s",
					Usage: "Set the sub store to operate on",
				},
				cli.StringFlag{
					Name:  "alias, a",
					Usage: "Set the name of the sub store",
				},
				cli.BoolFlag{
					Name:  "nogit",
					Usage: "Do not init git repo",
				},
			},
		},
		{
			Name:  "insert",
			Usage: "Insert new secret",
			Description: "" +
				"Insert new secret. Optionally, echo the secret back to the console during entry. " +
				"Or, optionally, the entry may be multiline. " +
				"Prompt before overwriting existing secret unless forced.",
			Before:       action.Initialized,
			Action:       action.Insert,
			BashComplete: action.Complete,
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
			Name:         "list",
			Usage:        "List secrets.",
			Aliases:      []string{"ls"},
			Before:       action.Initialized,
			Action:       action.List,
			BashComplete: action.Complete,
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
			Name:         "move",
			Aliases:      []string{"mv"},
			Usage:        "Renames or moves old-path to new-path, optionally forcefully, selectively reencrypting.",
			Before:       action.Initialized,
			Action:       action.Move,
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to move the secret and overwrite existing one",
				},
			},
		},
		{
			Name:        "mounts",
			Usage:       "Edit mounts",
			Description: "To manipulate gopass mounts",
			Before:      action.Initialized,
			Action:      action.MountsPrint,
			Subcommands: []cli.Command{
				{
					Name:        "add",
					Usage:       "Add mount",
					Description: "To add a new mounted sub store",
					Before:      action.Initialized,
					Action:      action.MountAdd,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "init, i",
							Usage: "Init the store with the given recipient key",
						},
					},
				},
				{
					Name:         "remove",
					Aliases:      []string{"rm"},
					Usage:        "Remove mount",
					Description:  "To remove a mounted sub store",
					Before:       action.Initialized,
					Action:       action.MountRemove,
					BashComplete: action.MountsComplete,
				},
			},
		},
		{
			Name:        "recipients",
			Usage:       "List Recipients",
			Description: "To show all recipients",
			Before:      action.Initialized,
			Action:      action.RecipientsPrint,
			Subcommands: []cli.Command{
				{
					Name:        "add",
					Usage:       "Add any number of Recipients",
					Description: "To add any number of recipients to a store",
					Before:      action.Initialized,
					Action:      action.RecipientsAdd,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
					},
				},
				{
					Name:         "remove",
					Aliases:      []string{"rm"},
					Usage:        "Remove any number of Recipients",
					Description:  "To remove any number of recipients from a store",
					Before:       action.Initialized,
					Action:       action.RecipientsRemove,
					BashComplete: action.RecipientsComplete,
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
			Name:  "show",
			Usage: "Show existing secret and optionally put its first line on the clipboard.",
			Description: "" +
				"Show existing secret and optionally put its first line on the clipboard. " +
				"If put on the clipboard, it will be cleared in 45 seconds.",
			Before:       action.Initialized,
			Action:       action.Show,
			BashComplete: action.Complete,
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
			},
		},
		{
			Name:  "templates",
			Usage: "List and edit secret templates.",
			Description: "" +
				"List existing templates in the password store and allow for editing " +
				"and creating them.",
			Before: action.Initialized,
			Action: action.TemplatesPrint,
			Subcommands: []cli.Command{
				{
					Name:         "show",
					Usage:        "Show a secret template.",
					Description:  "Dispaly an existing template",
					Aliases:      []string{"cat"},
					Before:       action.Initialized,
					Action:       action.TemplatePrint,
					BashComplete: action.TemplatesComplete,
				},
				{
					Name:         "edit",
					Usage:        "Edit secret templates.",
					Description:  "Edit an existing or new template",
					Aliases:      []string{"create", "new"},
					Before:       action.Initialized,
					Action:       action.TemplateEdit,
					BashComplete: action.TemplatesComplete,
				},
				{
					Name:         "remove",
					Aliases:      []string{"rm"},
					Usage:        "Remove secret templates.",
					Description:  "Remove an existing template",
					Before:       action.Initialized,
					Action:       action.TemplateRemove,
					BashComplete: action.TemplatesComplete,
				},
			},
		},
		{
			Name:        "unclip",
			Usage:       "Internal command to clear clipboard",
			Description: "Clear the clipboard if the content matches the checksum",
			Action:      action.Unclip,
			Hidden:      true,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "timeout",
					Usage: "Time to wait",
				},
			},
		},
		{
			Name:        "version",
			Usage:       "Print gopass version",
			Description: "Display version and build time information",
			Action:      action.Version,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
