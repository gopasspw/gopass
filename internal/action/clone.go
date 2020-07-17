package action

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Clone will fetch and mount a new password store from a git repo
func (s *Action) Clone(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("crypto") {
		ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))
	}
	if c.IsSet("sync") {
		ctx = backend.WithRCSBackendString(ctx, c.String("sync"))
	}

	if c.Args().Len() < 1 {
		return ExitError(ExitUsage, nil, "Usage: %s clone repo [mount]", s.Name)
	}

	repo := c.Args().Get(0)
	mount := ""
	if c.Args().Len() > 1 {
		mount = c.Args().Get(1)
	}

	path := c.String("path")

	return s.clone(ctx, repo, mount, path)
}

func rcsBackendOrDefault(ctx context.Context) backend.RCSBackend {
	if be := backend.GetRCSBackend(ctx); be != backend.Noop {
		return be
	}
	return backend.GitCLI
}

func (s *Action) clone(ctx context.Context, repo, mount, path string) error {
	if path == "" {
		path = config.PwStoreDir(mount)
	}
	inited, err := s.Store.Initialized(ctxutil.WithGitInit(ctx, false))
	if err != nil {
		return ExitError(ExitUnknown, err, "Failed to initialized stores: %s", err)
	}
	if mount == "" && inited {
		return ExitError(ExitAlreadyInitialized, nil, "Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
	}

	// make sure the parent directory exists
	if parentPath := filepath.Dir(path); !fsutil.IsDir(parentPath) {
		if err := os.MkdirAll(parentPath, 0700); err != nil {
			return ExitError(ExitUnknown, err, "Failed to create parent directory for clone: %s", err)
		}
	}

	// clone repo
	debug.Log("Cloning repo '%s' to '%s'", repo, path)
	if _, err := backend.CloneRCS(ctx, rcsBackendOrDefault(ctx), repo, path); err != nil {
		return ExitError(ExitGit, err, "failed to clone repo '%s' to '%s': %s", repo, path, err)
	}

	// add mount
	debug.Log("Mounting cloned repo '%s' at '%s'", path, mount)
	if err := s.cloneAddMount(ctx, mount, path); err != nil {
		return err
	}

	// save new mount in config file
	if err := s.cfg.Save(); err != nil {
		return ExitError(ExitIO, err, "Failed to update config: %s", err)
	}

	// try to init git config
	out.Green(ctx, "Configuring git repository ...")

	// ask for git config values
	username, email, err := s.cloneGetGitConfig(ctx, mount)
	if err != nil {
		return err
	}

	// initialize git config
	if err := s.Store.RCSInitConfig(ctx, mount, username, email); err != nil {
		debug.Log("Stacktrace: %+v\n", err)
		out.Error(ctx, "Failed to configure git: %s", err)
	}

	if mount != "" {
		mount = " " + mount
	}
	out.Green(ctx, "Your password store is ready to use! Have a look around: `%s list%s`\n", s.Name, mount)

	return nil
}

func (s *Action) cloneAddMount(ctx context.Context, mount, path string) error {
	if mount == "" {
		return nil
	}

	inited, err := s.Store.Initialized(ctx)
	if err != nil {
		return ExitError(ExitUnknown, err, "Failed to initialize store: %s", err)
	}
	if !inited {
		return ExitError(ExitNotInitialized, nil, "Root-Store is not initialized. Clone or init root store first")
	}
	if err := s.Store.AddMount(ctx, mount, path); err != nil {
		return ExitError(ExitMount, err, "Failed to add mount: %s", err)
	}
	out.Green(ctx, "Mounted password store %s at mount point `%s` ...", path, mount)
	return nil
}

func (s *Action) cloneGetGitConfig(ctx context.Context, name string) (string, string, error) {
	// for convenience, set defaults to user-selected values from available private keys
	// NB: discarding returned error since this is merely a best-effort look-up for convenience
	username, email, _ := cui.AskForGitConfigUser(ctx, s.Store.Crypto(ctx, name))
	if username == "" {
		username = termio.DetectName(ctx, nil)
		var err error
		username, err = termio.AskForString(ctx, color.CyanString("Please enter a user name for password store git config"), username)
		if err != nil {
			return "", "", ExitError(ExitIO, err, "Failed to read user input: %s", err)
		}
	}
	if email == "" {
		email = termio.DetectEmail(ctx, nil)
		var err error
		email, err = termio.AskForString(ctx, color.CyanString("Please enter an email address for password store git config"), email)
		if err != nil {
			return "", "", ExitError(ExitIO, err, "Failed to read user input: %s", err)
		}
	}
	return username, email, nil
}
