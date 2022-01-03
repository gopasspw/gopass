package action

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	si "github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// RCSInit initializes a git repo including basic configuration.
func (s *Action) RCSInit(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	un := termio.DetectName(c.Context, c)
	ue := termio.DetectEmail(c.Context, c)
	ctx = backend.WithStorageBackendString(ctx, c.String("storage"))

	// default to git.
	if !backend.HasStorageBackend(ctx) {
		ctx = backend.WithStorageBackend(ctx, backend.GitFS)
	}

	if err := s.rcsInit(ctx, store, un, ue); err != nil {
		return ExitError(ExitGit, err, "failed to initialize git: %s", err)
	}
	return nil
}

func (s *Action) rcsInit(ctx context.Context, store, un, ue string) error {
	be := backend.GetStorageBackend(ctx)
	if err := s.Store.RCSStatus(ctx, store); errors.Is(err, backend.ErrNotSupported) {
		// no git support, skip
		return nil
	}

	bn := backend.StorageBackendName(be)
	userName, userEmail := s.getUserData(ctx, store, un, ue)
	if err := s.Store.RCSInit(ctx, store, userName, userEmail); err != nil {
		if errors.Is(err, backend.ErrNotSupported) {
			debug.Log("RCSInit not supported for backend %s", bn)
			return nil
		}
		if gtv := os.Getenv("GPG_TTY"); gtv == "" {
			out.Printf(ctx, "Git initialization failed. You may want to try to 'export GPG_TTY=$(tty)' and start over.")
		}
		return fmt.Errorf("failed to run git init: %w", err)
	}

	out.Printf(ctx, "Initialized git repository (%s) for %s / %s...", bn, un, ue)
	return nil
}

func (s *Action) getUserData(ctx context.Context, store, name, email string) (string, string) {
	if name != "" && email != "" {
		debug.Log("Username: %s, Email: %s (provided)", name, email)
		return name, email
	}

	// for convenience, set defaults to user-selected values from available private keys.
	// NB: discarding returned error since this is merely a best-effort look-up for convenience.
	userName, userEmail, _ := cui.AskForGitConfigUser(ctx, s.Store.Crypto(ctx, store))

	if name == "" {
		if userName == "" {
			userName = termio.DetectName(ctx, nil)
		}

		var err error
		name, err = termio.AskForString(ctx, color.CyanString("Please enter a user name for password store git config"), userName)
		if err != nil {
			out.Errorf(ctx, "Failed to ask for user input: %s", err)
		}
	}
	if email == "" {
		if userEmail == "" {
			userEmail = termio.DetectEmail(ctx, nil)
		}

		var err error
		email, err = termio.AskForString(ctx, color.CyanString("Please enter an email address for password store git config"), userEmail)
		if err != nil {
			out.Errorf(ctx, "Failed to ask for user input: %s", err)
		}
	}

	debug.Log("Username: %s, Email: %s (detected)", name, email)
	return name, email
}

// RCSAddRemote adds a new git remote.
func (s *Action) RCSAddRemote(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	remote := c.Args().Get(0)
	url := c.Args().Get(1)

	if remote == "" || url == "" {
		return ExitError(ExitUsage, nil, "Usage: %s git remote add <REMOTE> <URL>", s.Name)
	}

	return s.Store.RCSAddRemote(ctx, store, remote, url)
}

// RCSRemoveRemote removes a git remote.
func (s *Action) RCSRemoveRemote(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	remote := c.Args().Get(0)

	if remote == "" {
		return ExitError(ExitUsage, nil, "Usage: %s git remote rm <REMOTE>", s.Name)
	}

	return s.Store.RCSRemoveRemote(ctx, store, remote)
}

// RCSPull pulls from a git remote.
func (s *Action) RCSPull(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	origin := c.Args().Get(0)
	branch := c.Args().Get(1)

	return s.Store.RCSPull(ctx, store, origin, branch)
}

// RCSPush pushes to a git remote.
func (s *Action) RCSPush(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	origin := c.Args().Get(0)
	branch := c.Args().Get(1)

	if err := s.Store.RCSPush(ctx, store, origin, branch); err != nil {
		if errors.Is(err, si.ErrGitNoRemote) {
			out.Noticef(ctx, "No Git remote. Not pushing")
			return nil
		}
		return ExitError(ExitGit, err, "Failed to push to remote")
	}
	out.OKf(ctx, "Pushed to git remote")
	return nil
}

// RCSStatus prints the rcs status.
func (s *Action) RCSStatus(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")

	return s.Store.RCSStatus(ctx, store)
}
