package action

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	git "github.com/justwatchcom/gopass/backend/git/cli"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termio"
	"github.com/urfave/cli"
)

// Clone will fetch and mount a new password store from a git repo
func (s *Action) Clone(ctx context.Context, c *cli.Context) error {
	if len(c.Args()) < 1 {
		return exitError(ctx, ExitUsage, nil, "Usage: %s clone repo [mount]", s.Name)
	}

	repo := c.Args()[0]
	mount := ""
	if len(c.Args()) > 1 {
		mount = c.Args()[1]
	}

	path := c.String("path")

	return s.clone(ctx, repo, mount, path)
}

func (s *Action) clone(ctx context.Context, repo, mount, path string) error {
	if path == "" {
		path = config.PwStoreDir(mount)
	}
	if mount == "" && s.Store.Initialized() {
		return exitError(ctx, ExitAlreadyInitialized, nil, "Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
	}

	// clone repo
	if _, err := git.Clone(ctx, s.gpg.Binary(), repo, path); err != nil {
		return exitError(ctx, ExitGit, err, "failed to clone repo '%s' to '%s'", repo, path)
	}

	// add mount
	if mount != "" {
		if !s.Store.Initialized() {
			return exitError(ctx, ExitNotInitialized, nil, "Root-Store is not initialized. Clone or init root store first")
		}
		if err := s.Store.AddMount(ctx, mount, path); err != nil {
			return exitError(ctx, ExitMount, err, "Failed to add mount: %s", err)
		}
		fmt.Printf("Mounted password store %s at mount point `%s` ...\n", path, mount)
	}

	// save new mount in config file
	if err := s.cfg.Save(); err != nil {
		return exitError(ctx, ExitIO, err, "Failed to update config: %s", err)
	}

	// try to init git config
	out.Green(ctx, "Configuring git repository ...")
	sk, err := s.askForPrivateKey(ctx, color.CyanString("Please select a key for signing Git Commits"))
	if err != nil {
		out.Red(ctx, "Failed to ask for signing key: %s", err)
	}

	// ask for git config values
	username, email, err := s.cloneGetGitConfig(ctx)
	if err != nil {
		return err
	}

	// initialize git config
	if err := s.Store.GitInitConfig(ctx, mount, sk, username, email); err != nil {
		out.Debug(ctx, "Stacktrace: %+v\n", err)
		out.Red(ctx, "Failed to configure git: %s", err)
	}

	if mount != "" {
		mount = " " + mount
	}
	out.Green(ctx, "Your password store is ready to use! Have a look around: `%s list%s`\n", s.Name, mount)

	return nil
}

func (s *Action) cloneGetGitConfig(ctx context.Context) (string, string, error) {
	// for convenience, set defaults to user-selected values from available private keys
	// NB: discarding returned error since this is merely a best-effort look-up for convenience
	username, email, _ := s.askForGitConfigUser(ctx)
	if username == "" {
		var err error
		username, err = termio.AskForString(ctx, color.CyanString("Please enter a user name for password store git config"), username)
		if err != nil {
			return "", "", exitError(ctx, ExitIO, err, "Failed to read user input: %s", err)
		}
	}
	if email == "" {
		var err error
		email, err = termio.AskForString(ctx, color.CyanString("Please enter an email address for password store git config"), email)
		if err != nil {
			return "", "", exitError(ctx, ExitIO, err, "Failed to read user input: %s", err)
		}
	}
	return username, email, nil
}
