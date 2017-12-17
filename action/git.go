package action

import (
	"context"
	"os"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Git runs git commands inside the store or mounts
func (s *Action) Git(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	recurse := true
	if c.IsSet("no-recurse") {
		recurse = !c.Bool("no-recurse")
	}
	force := c.Bool("force")

	if err := s.Store.Git(ctxutil.WithVerbose(ctx, true), store, recurse, force, c.Args()...); err != nil {
		return exitError(ctx, ExitGit, err, "git operation failed: %s", err)
	}
	return nil
}

// GitInit initializes a git repo including basic configuration
func (s *Action) GitInit(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	sk := c.String("sign-key")

	if err := s.gitInit(ctx, store, sk); err != nil {
		return exitError(ctx, ExitGit, err, "failed to initialize git: %s", err)
	}
	return nil
}

func (s *Action) gitInit(ctx context.Context, store, sk string) error {
	out.Green(ctx, "Initializing git repository ...")
	if sk == "" {
		s, err := s.askForPrivateKey(ctx, color.CyanString("Please select a key for signing Git Commits"))
		if err == nil {
			sk = s
		}
	}

	// for convenience, set defaults to user-selected values from available private keys
	// NB: discarding returned error since this is merely a best-effort look-up for convenience
	userName, userEmail, _ := s.askForGitConfigUser(ctx)

	if userName == "" {
		var err error
		userName, err = s.askForString(ctx, color.CyanString("Please enter a user name for password store git config"), userName)
		if err != nil {
			return errors.Wrapf(err, "failed to ask for user input")
		}
	}
	if userEmail == "" {
		var err error
		userEmail, err = s.askForString(ctx, color.CyanString("Please enter an email address for password store git config"), userEmail)
		if err != nil {
			return errors.Wrapf(err, "failed to ask for user input")
		}
	}

	if err := s.Store.GitInit(ctx, store, sk, userName, userEmail); err != nil {
		if gtv := os.Getenv("GPG_TTY"); gtv == "" {
			out.Yellow(ctx, "Git initialization failed. You may want to try to 'export GPG_TTY=$(tty)' and start over.")
		}
		return errors.Wrapf(err, "failed to run git init")
	}

	out.Green(ctx, "Git initialized")
	return nil
}
