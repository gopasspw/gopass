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
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/protect"
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
	protect.Pledge("stdio rpath tty proc exec")
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

	// always use symbols
	ctx = ctxutil.WithUseSymbols(ctx, cfg.Root.UseSymbols)

	// never use color
	ctx = ctxutil.WithNoColor(ctx, cfg.Root.NoColor)

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
	if nc := os.Getenv("GOPASS_NOCOLOR"); nc == "true" || ctxutil.IsNoColor(ctx) {
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
			Minor: 6,
			Patch: 1,
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
			out.Red(ctx, "Failed to save config: %s", err)
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
		if err := action.Initialized(withGlobalFlags(ctx, c), c); err != nil {
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
					Usage: "Detect leaked passwords (ALPHA)",
					Description: "" +
						"This command will decrypt all secrets and check the passwords against the public " +
						"havibeenpwned.com dumps. " +
						"To use this feature you need to download the dumps from https://haveibeenpwned.com/passwords first. This is a very expensive operation for advanced users",
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
						return action.BinaryCat(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.Complete,
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
						return action.BinarySum(withGlobalFlags(ctx, c), c)
					},
					BashComplete: action.Complete,
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
				Name:   "zsh",
				Usage:  "Source for auto completion in zsh",
				Action: action.CompletionZSH,
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
				"It will also handle copying secrets to different sub stores.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
			Name:    "create",
			Aliases: []string{"new"},
			Usage:   "Easy creation of new secrets",
			Description: "" +
				"This command starts a wizard to aid in creation of new secrets.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Create(withGlobalFlags(ctx, c), c)
			},
		},
		{
			Name:  "delete",
			Usage: "Remove secrets",
			Description: "" +
				"This command removes secrets. It can work recursivly on folders. " +
				"Recursing across stores is purposefully not supported.",
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
			BashComplete: action.Complete,
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
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "clip, c",
					Usage: "Copy the password into the clipboard",
				},
			},
		},
		{
			Name:   "fix",
			Usage:  "Upgrade secrets",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
			Action: func(c *cli.Context) error {
				return action.Fix(withGlobalFlags(ctx, c), c)
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
			Name:  "fsck",
			Usage: "Check inconsistencies (ALPHA)",
			Description: "" +
				"Check all mounted password stores for know issues and inconsistencies, like " +
				"wrong file persmissions or missing / extra recipients.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
			BashComplete: action.Complete,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "print, p",
					Usage: "Print the generated password to the tmerinal",
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
			BashComplete: action.Complete,
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
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
				cli.BoolFlag{
					Name:  "nogit",
					Usage: "Do not init git repo",
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
			Usage:   "Move secrets from one location to another",
			Description: "" +
				"This command moves a secret from one path to another. This works even " +
				"across different sub stores.",
			Before: func(c *cli.Context) error { return action.Initialized(withGlobalFlags(ctx, c), c) },
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
					Usage:   "Mount an password store",
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
				cli.BoolFlag{
					Name:  "password, o",
					Usage: "Dispaly only the password",
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
