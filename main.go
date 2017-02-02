package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/action"
	"github.com/urfave/cli"
)

const (
	name = "gopass"
)

var (
	// Version is the released version of gopass
	Version string
	// BuildTime is the time the binary was built
	BuildTime string
	// Commit is the git hash the binary was built from
	Commit string
)

type errorWriter struct {
	out io.Writer
}

func (e errorWriter) Write(p []byte) (int, error) {
	return e.out.Write([]byte("\n" + color.RedString("Error: "+string(p))))
}

func main() {
	cli.ErrWriter = errorWriter{
		out: os.Stderr,
	}

	cli.VersionPrinter = func(c *cli.Context) {
		buildtime := ""
		if bt, err := time.Parse("2006-01-02T15:04:05-0700", BuildTime); err == nil {
			buildtime = bt.Format("2006-01-02 15:04:05")
		}
		if Version == "" {
			Version = "HEAD"
		}
		if Commit == "" {
			Commit = "n/a"
		}
		fmt.Printf("%s %s (%s %s) %s\n",
			name,
			Version,
			Commit,
			buildtime,
			runtime.Version(),
		)
	}

	action := action.New(Version)
	app := cli.NewApp()

	app.Name = name
	app.Version = Version
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
			Usage: "Copy the secret into the clipboard",
		},
	}

	app.Commands = []cli.Command{
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
			Name:        "git",
			Usage:       "Do git things",
			Description: "If the password store is a git repository, execute a git command specified by git-command-args.",
			Before:      action.Initialized,
			Action:      action.Git,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "store",
					Usage: "Store to operate on",
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
					Usage: "Overwrite any existing secret",
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
							Usage: "Init the store with the given recpient key",
						},
					},
				},
				{
					Name:         "remove",
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
			Usage: "Show existing secret and optionally put it on the clipboard.",
			Description: "" +
				"Show existing secret and optionally put it on the clipboard. " +
				"If put on the clipboard, it will be cleared in 45 seconds.",
			Before:       action.Initialized,
			Action:       action.Show,
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the secret into the clipboard",
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
