package action

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// GitInit initializes a git repo including basic configuration
func (s *Action) GitInit(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	un := termio.DetectName(c.Context, c)
	ue := termio.DetectEmail(c.Context, c)
	ctx = backend.WithRCSBackendString(ctx, c.String("rcs"))

	// default to git
	if !backend.HasRCSBackend(ctx) {
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	}

	if err := s.rcsInit(ctx, store, un, ue); err != nil {
		return ExitError(ExitGit, err, "failed to initialize git: %s", err)
	}
	return nil
}

func (s *Action) rcsInit(ctx context.Context, store, un, ue string) error {
	bn := backend.RCSBackendName(backend.GetRCSBackend(ctx))
	out.Green(ctx, "Initializing git repository (%s) for %s / %s...", bn, un, ue)

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

func (s *Action) getUserData(ctx context.Context, store, name, email string) (string, string) {
	if name != "" && email != "" {
		debug.Log("Username: %s, Email: %s (provided)", name, email)
		return name, email
	}

	// for convenience, set defaults to user-selected values from available private keys
	// NB: discarding returned error since this is merely a best-effort look-up for convenience
	userName, userEmail, _ := cui.AskForGitConfigUser(ctx, s.Store.Crypto(ctx, store))

	if name == "" {
		if userName == "" {
			userName = termio.DetectName(ctx, nil)
		}
		var err error
		name, err = termio.AskForString(ctx, color.CyanString("Please enter a user name for password store git config"), userName)
		if err != nil {
			out.Error(ctx, "Failed to ask for user input: %s", err)
		}
	}
	if email == "" {
		if userEmail == "" {
			userEmail = termio.DetectEmail(ctx, nil)
		}
		var err error
		email, err = termio.AskForString(ctx, color.CyanString("Please enter an email address for password store git config"), userEmail)
		if err != nil {
			out.Error(ctx, "Failed to ask for user input: %s", err)
		}
	}

	debug.Log("Username: %s, Email: %s (detected)", name, email)
	return name, email
}

// GitAddRemote adds a new git remote
func (s *Action) GitAddRemote(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	remote := c.Args().Get(0)
	url := c.Args().Get(1)

	if remote == "" || url == "" {
		return ExitError(ExitUsage, nil, "Usage: %s git remote add <REMOTE> <URL>", s.Name)
	}

	return s.Store.GitAddRemote(ctx, store, remote, url)
}

// GitRemoveRemote removes a git remote
func (s *Action) GitRemoveRemote(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	remote := c.Args().Get(0)

	if remote == "" {
		return ExitError(ExitUsage, nil, "Usage: %s git remote rm <REMOTE>", s.Name)
	}

	return s.Store.GitRemoveRemote(ctx, store, remote)
}

// GitPull pulls from a git remote
func (s *Action) GitPull(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	origin := c.Args().Get(0)
	branch := c.Args().Get(1)

	if origin == "" {
		origin = "origin"
	}
	if branch == "" {
		branch = "master"
	}
	return s.Store.GitPull(ctx, store, origin, branch)
}

// GitPush pushes to a git remote
func (s *Action) GitPush(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	origin := c.Args().Get(0)
	branch := c.Args().Get(1)

	if origin == "" || branch == "" {
		return ExitError(ExitUsage, nil, "Usage: %s git push <ORIGIN> <BRANCH>", s.Name)
	}
	return s.Store.GitPush(ctx, store, origin, branch)
}

// GitStatus prints the rcs status
func (s *Action) GitStatus(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")

	return s.Store.GitStatus(ctx, store)
}
