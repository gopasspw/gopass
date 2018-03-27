package action

import (
	"context"
	"os"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/cui"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/termio"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// GitInit initializes a git repo including basic configuration
func (s *Action) GitInit(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	un := c.String("username")
	ue := c.String("useremail")
	ctx = backend.WithRCSBackendString(ctx, c.String("rcs"))

	// default to git
	if !backend.HasRCSBackend(ctx) {
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	}

	if err := s.rcsInit(ctx, store, un, ue); err != nil {
		return ExitError(ctx, ExitGit, err, "failed to initialize git: %s", err)
	}
	return nil
}

func (s *Action) rcsInit(ctx context.Context, store, un, ue string) error {
	bn := backend.RCSBackendName(backend.GetRCSBackend(ctx))
	out.Green(ctx, "Initializing git repository (%s) ...", bn)

	userName, userEmail := s.getUserData(ctx, store, un, ue)
	if err := s.Store.GitInit(ctx, store, userName, userEmail); err != nil {
		if gtv := os.Getenv("GPG_TTY"); gtv == "" {
			out.Yellow(ctx, "Git initialization failed. You may want to try to 'export GPG_TTY=$(tty)' and start over.")
		}
		return errors.Wrapf(err, "failed to run git init")
	}

	out.Green(ctx, "Git initialized")
	return nil
}

func (s *Action) getUserData(ctx context.Context, store, un, ue string) (string, string) {
	if un != "" && ue != "" {
		return un, ue
	}

	// for convenience, set defaults to user-selected values from available private keys
	// NB: discarding returned error since this is merely a best-effort look-up for convenience
	userName, userEmail, _ := cui.AskForGitConfigUser(ctx, s.Store.Crypto(ctx, store), store)

	if userName == "" {
		var err error
		userName, err = termio.AskForString(ctx, color.CyanString("Please enter a user name for password store git config"), userName)
		if err != nil {
			out.Red(ctx, "Failed to ask for user input: %s", err)
		}
	}
	if userEmail == "" {
		var err error
		userEmail, err = termio.AskForString(ctx, color.CyanString("Please enter an email address for password store git config"), userEmail)
		if err != nil {
			out.Red(ctx, "Failed to ask for user input: %s", err)
		}
	}

	return userName, userEmail
}

// GitAddRemote adds a new git remote
func (s *Action) GitAddRemote(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	remote := c.Args().Get(0)
	url := c.Args().Get(1)

	if remote == "" || url == "" {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s git remote add <REMOTE> <URL>", s.Name)
	}

	return s.Store.GitAddRemote(ctx, store, remote, url)
}

// GitPull pulls from a git remote
func (s *Action) GitPull(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	origin := c.Args().Get(0)
	branch := c.Args().Get(1)

	if origin == "" || branch == "" {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s git pull <ORIGIN> <BRANCH>", s.Name)
	}
	return s.Store.GitPull(ctx, store, origin, branch)
}

// GitPush pushes to a git remote
func (s *Action) GitPush(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	origin := c.Args().Get(0)
	branch := c.Args().Get(1)

	if origin == "" || branch == "" {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s git push <ORIGIN> <BRANCH>", s.Name)
	}
	return s.Store.GitPush(ctx, store, origin, branch)
}
