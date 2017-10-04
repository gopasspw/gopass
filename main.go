package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/action"
	"github.com/justwatchcom/gopass/backend/gpg"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/ctxutil"
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

func main() {
	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	// try to read config (if it exists)
	cfg := config.Load()

	// autosync
	ctx = sub.WithAutoSync(ctx, cfg.Root.AutoSync)

	// always trust
	ctx = gpg.WithAlwaysTrust(ctx, true)

	// ask for more
	ctx = ctxutil.WithAskForMore(ctx, cfg.Root.AskForMore)

	// clipboard timeout
	ctx = ctxutil.WithClipTimeout(ctx, cfg.Root.ClipTimeout)

	// no confirm
	ctx = ctxutil.WithNoConfirm(ctx, cfg.Root.NoConfirm)

	// no pager
	ctx = ctxutil.WithNoPager(ctx, cfg.Root.NoPager)

	// show safe content
	ctx = ctxutil.WithShowSafeContent(ctx, cfg.Root.SafeContent)

	// check recipients conflicts with always trust, make sure it's not enabled
	// when always trust is
	if gpg.IsAlwaysTrust(ctx) {
		ctx = sub.WithCheckRecipients(ctx, false)
	}

	// debug flag
	if gdb := os.Getenv("GOPASS_DEBUG"); gdb == "true" {
		ctx = ctxutil.WithDebug(ctx, true)
	}

	// need this override for our integration tests
	if nc := os.Getenv("GOPASS_NOCOLOR"); nc == "true" {
		color.NoColor = true
		ctx = ctxutil.WithColor(ctx, false)
	}

	// only emit color codes when stdout is a terminal
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		color.NoColor = true
		ctx = ctxutil.WithColor(ctx, false)
		ctx = ctxutil.WithTerminal(ctx, false)
		ctx = ctxutil.WithInteractive(ctx, false)
	}

	// reading from stdin?
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice == 0 {
		ctx = ctxutil.WithInteractive(ctx, false)
		ctx = ctxutil.WithStdin(ctx, true)
	}

	// disable colored output on windows since cmd.exe doesn't support ANSI color
	// codes. Other terminal may do, but until we can figure that out better
	// disable this for all terms on this platform
	if runtime.GOOS == "windows" {
		color.NoColor = true
		ctx = ctxutil.WithColor(ctx, false)
	}

	cli.ErrWriter = errorWriter{
		out: colorable.NewColorableStderr(),
	}

	sv, err := semver.Parse(version)
	if err != nil {
		sv = semver.Version{
			Major: 1,
			Minor: 3,
			Patch: 3,
			Pre: []semver.PRVersion{
				semver.PRVersion{VersionStr: "git"},
			},
			Build: []string{"HEAD"},
		}
	}

	// only update version field in config, if it's older than this build
	csv, err := semver.Parse(cfg.Version)
	if err != nil || csv.LT(sv) {
		cfg.Version = sv.String()
		if err := cfg.Save(); err != nil {
			fmt.Println(color.RedString("Failed to save config: %s", err))
		}
	}

	cli.VersionPrinter = makeVersionPrinter(sv)

	action := action.New(ctx, cfg, sv)

	// set some action callbacks
	if !cfg.Root.AutoImport {
		ctx = sub.WithImportFunc(ctx, action.AskForKeyImport)
	}
	if !cfg.Root.NoConfirm {
		ctx = sub.WithRecipientFunc(ctx, action.ConfirmRecipients)
	}
	ctx = sub.WithFsckFunc(ctx, action.AskForConfirmation)

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
		if err := action.Initialized(ctx, c); err != nil {
			return err
		}

		if c.Args().Present() {
			return action.Show(withGlobalFlags(ctx, c), c)
		}
		return action.List(withGlobalFlags(ctx, c), c)
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "yes",
			Usage: "Assume yes on all yes/no questions or use the default on all others",
		},
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
			Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
					Name:   "hibp",
					Usage:  "Check all secrets against the public haveibeenpwned.com dumps",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.HIBP(withGlobalFlags(ctx, c), c)
					},
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
			Name:    "binary",
			Usage:   "Work with binary blobs",
			Aliases: []string{"bin"},
			Subcommands: []cli.Command{
				{
					Name:   "cat",
					Usage:  "Print content of a secret to stdout or insert from stdin",
					Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.BinaryCat(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.Complete,
				},
				{
					Name:    "sum",
					Usage:   "Compute the SHA256 sum of a decoded secret",
					Aliases: []string{"sha", "sha256"},
					Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.BinarySum(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.Complete,
				},
				{
					Name:    "copy",
					Usage:   "Copy files from or to the password store",
					Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Aliases: []string{"cp"},
					Action: func(c *cli.Context) error {
						return action.BinaryCopy(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.Complete,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Force to move the secret and overwrite existing one",
						},
					},
				},
				{
					Name:    "move",
					Usage:   "Move files from or to the password store",
					Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Aliases: []string{"mv"},
					Action: func(c *cli.Context) error {
						return action.BinaryMove(withGlobalFlags(ctx, c), c)
					},
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
			Action: func(c *cli.Context) error {
				return action.Clone(withGlobalFlags(ctx, c), c)
			},
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
			Usage:   "Copies old-path to new-path, optionally forcefully, selectively reencrypting.",
			Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Copy(withGlobalFlags(ctx, c), c)
			},
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Force to copy the secret and overwrite existing one",
				},
			},
		},
		{
			Name:    "delete",
			Usage:   "Remove existing secret or directory, optionally forcefully.",
			Aliases: []string{"remove", "rm"},
			Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Delete(withGlobalFlags(ctx, c), c)
			},
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
			Name:   "edit",
			Usage:  "Insert a new secret or edit an existing secret using $EDITOR.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Edit(withGlobalFlags(ctx, c), c)
			},
			Aliases:      []string{"set"},
			BashComplete: action.Complete,
		},
		{
			Name:   "find",
			Usage:  "List secrets that match the search term.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Find(withGlobalFlags(ctx, c), c)
			},
			Aliases:      []string{"search"},
			BashComplete: action.Complete,
		},
		{
			Name:        "fsck",
			Usage:       "Check store integrity",
			Description: "Check integrity of all stores",
			Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Fsck(withGlobalFlags(ctx, c), c)
			},
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
				"It will replace only the first line of an existing file with a new password.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Generate(withGlobalFlags(ctx, c), c)
			},
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
					Name:  "edit, e",
					Usage: "Open secret for editing after generating a password",
				},
				cli.BoolFlag{
					Name:   "no-symbols, n",
					Usage:  "Do not include symbols in the password",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:  "symbols, s",
					Usage: "Use symbols in the password",
				},
				cli.BoolFlag{
					Name:  "xkcd, x",
					Usage: "Use multiple random english words as password, separated by space",
				},
				cli.BoolFlag{
					Name:  "xkcdo, xo",
					Usage: "Use multiple random english words as password, no separator but CamelCased",
				},
			},
		},
		{
			Name:        "jsonapi",
			Usage:       "Run gopass as jsonapi e.g. for browser plugins",
			Description: "Setup and run gopass as native messaging hosts, e.g. for browser plugins.",
			Subcommands: []cli.Command{
				{
					Name:        "listen",
					Usage:       "Listen and respond to messages via stdin/stdout",
					Description: "Gopass is started in listen mode from browser plugins using a wrapper specified in native messaging host manifests",
					Action: func(c *cli.Context) error {
						return action.JSONAPI(withGlobalFlags(ctx, c), c)
					},
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
						cli.BoolFlag{
							Name:  "global",
							Usage: "Install for all users, requires superuser rights",
						},
						cli.StringFlag{
							Name:  "libpath",
							Usage: "Library path for global installation on linux. Default is /usr/lib",
						},
						cli.BoolFlag{
							Name:  "print-only",
							Usage: "only print installation summary but do not actually create any files",
						},
					},
				},
			},
		},
		{
			Name:        "totp",
			Usage:       "Generate time based token from stored secret",
			Description: "Tries to parse the saved string as a time-based one-time password secret and generate a token based on the current time",
			Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.TOTP(withGlobalFlags(ctx, c), c)
			},
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
			Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Git(withGlobalFlags(ctx, c), c)
			},
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
					},
				},
			},
		},
		{
			Name:   "grep",
			Usage:  "Search for secrets files containing search-string when decrypted.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Grep(withGlobalFlags(ctx, c), c)
			},
		},
		{
			Name:  "init",
			Usage: "Initialize new password storage and use gpg-id for encryption.",
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
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Insert(withGlobalFlags(ctx, c), c)
			},
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
			Name:    "list",
			Usage:   "List secrets.",
			Aliases: []string{"ls"},
			Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.List(withGlobalFlags(ctx, c), c)
			},
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
			Name:    "move",
			Aliases: []string{"mv"},
			Usage:   "Renames or moves old-path to new-path, optionally forcefully, selectively reencrypting.",
			Before:  func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Move(withGlobalFlags(ctx, c), c)
			},
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
			Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.MountsPrint(withGlobalFlags(ctx, c), c)
			},
			Subcommands: []cli.Command{
				{
					Name:        "add",
					Usage:       "Add mount",
					Description: "To add a new mounted sub store",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
					Name:        "remove",
					Aliases:     []string{"rm"},
					Usage:       "Remove mount",
					Description: "To remove a mounted sub store",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.MountRemove(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.MountsComplete,
				},
			},
		},
		{
			Name:        "recipients",
			Usage:       "List Recipients",
			Description: "To show all recipients",
			Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.RecipientsPrint(withGlobalFlags(ctx, c), c)
			},
			Subcommands: []cli.Command{
				{
					Name:        "add",
					Usage:       "Add any number of Recipients",
					Description: "To add any number of recipients to a store",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
					Name:        "remove",
					Aliases:     []string{"rm"},
					Usage:       "Remove any number of Recipients",
					Description: "To remove any number of recipients from a store",
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
			Name:  "show",
			Usage: "Show existing secret and optionally put its first line on the clipboard.",
			Description: "" +
				"Show existing secret and optionally put its first line on the clipboard. " +
				"If put on the clipboard, it will be cleared in 45 seconds.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Show(withGlobalFlags(ctx, c), c)
			},
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
			Name:  "sync",
			Usage: "Sync all local stores with their remotes (if any)",
			Description: "" +
				"Sync all local stores with their git remotes, if any, and check" +
				"any possibly affected gpg keys.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Sync(withGlobalFlags(ctx, c), c)

			},
		},
		{
			Name:  "templates",
			Usage: "List and edit secret templates.",
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
					Description: "Dispaly an existing template",
					Aliases:     []string{"cat"},
					Before:      func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
					Action: func(c *cli.Context) error {
						return action.TemplatePrint(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.TemplatesComplete,
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
					BashComplete: action.TemplatesComplete,
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
					BashComplete: action.TemplatesComplete,
				},
			},
		},
		{
			Name:        "unclip",
			Usage:       "Internal command to clear clipboard",
			Description: "Clear the clipboard if the content matches the checksum",
			Action: func(c *cli.Context) error {
				return action.Unclip(withGlobalFlags(ctx, c), c)
			},
			Hidden: true,
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
			Action: func(c *cli.Context) error {
				return action.Version(withGlobalFlags(ctx, c), c)
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func makeVersionPrinter(sv semver.Version) func(c *cli.Context) {
	return func(c *cli.Context) {
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
}

type errorWriter struct {
	out io.Writer
}

func (e errorWriter) Write(p []byte) (int, error) {
	return e.out.Write([]byte("\n" + color.RedString("Error: %s", p)))
}

func withGlobalFlags(ctx context.Context, c *cli.Context) context.Context {
	if c.GlobalBool("yes") {
		ctx = ctxutil.WithAlwaysYes(ctx, true)
	}
	return ctx
}
