package action

import (
	"context"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/backend/crypto/xc"
	gitcli "github.com/justwatchcom/gopass/backend/sync/git/cli"
	"github.com/justwatchcom/gopass/backend/sync/git/gogit"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termio"
	"github.com/urfave/cli"
)

// Clone will fetch and mount a new password store from a git repo
func (s *Action) Clone(ctx context.Context, c *cli.Context) error {
	ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))
	ctx = backend.WithSyncBackendString(ctx, c.String("sync"))

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
	if mount == "" && s.Store.Initialized(ctx) {
		return exitError(ctx, ExitAlreadyInitialized, nil, "Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
	}

	// clone repo
	switch backend.GetSyncBackend(ctx) {
	case backend.GoGit:
		if _, err := gogit.Clone(ctx, repo, path); err != nil {
			return exitError(ctx, ExitGit, err, "failed to clone repo '%s' to '%s'", repo, path)
		}
	case backend.GitCLI:
		fallthrough
	default:
		if _, err := gitcli.Clone(ctx, repo, path); err != nil {
			return exitError(ctx, ExitGit, err, "failed to clone repo '%s' to '%s'", repo, path)
		}
	}

	// detect crypto backend based on cloned repo
	ctx = backend.WithCryptoBackend(ctx, detectCryptoBackend(ctx, path))

	// add mount
	if mount != "" {
		if !s.Store.Initialized(ctx) {
			return exitError(ctx, ExitNotInitialized, nil, "Root-Store is not initialized. Clone or init root store first")
		}
		if err := s.Store.AddMount(ctx, mount, path); err != nil {
			return exitError(ctx, ExitMount, err, "Failed to add mount: %s", err)
		}
		out.Green(ctx, "Mounted password store %s at mount point `%s` ...", path, mount)
		s.cfg.Mounts[mount].CryptoBackend = backend.CryptoBackendName(backend.GetCryptoBackend(ctx))
		s.cfg.Mounts[mount].SyncBackend = backend.SyncBackendName(backend.GetSyncBackend(ctx))
		s.cfg.Mounts[mount].StoreBackend = backend.StoreBackendName(backend.GetStoreBackend(ctx))
	} else {
		s.cfg.Root.CryptoBackend = backend.CryptoBackendName(backend.GetCryptoBackend(ctx))
		s.cfg.Root.SyncBackend = backend.SyncBackendName(backend.GetSyncBackend(ctx))
		s.cfg.Root.StoreBackend = backend.StoreBackendName(backend.GetStoreBackend(ctx))
	}

	// save new mount in config file
	if err := s.cfg.Save(); err != nil {
		return exitError(ctx, ExitIO, err, "Failed to update config: %s", err)
	}

	// try to init git config
	out.Green(ctx, "Configuring git repository ...")

	// ask for git config values
	username, email, err := s.cloneGetGitConfig(ctx, mount)
	if err != nil {
		return err
	}

	// initialize git config
	if err := s.Store.GitInitConfig(ctx, mount, username, email); err != nil {
		out.Debug(ctx, "Stacktrace: %+v\n", err)
		out.Red(ctx, "Failed to configure git: %s", err)
	}

	if mount != "" {
		mount = " " + mount
	}
	out.Green(ctx, "Your password store is ready to use! Have a look around: `%s list%s`\n", s.Name, mount)

	return nil
}

func (s *Action) cloneGetGitConfig(ctx context.Context, name string) (string, string, error) {
	// for convenience, set defaults to user-selected values from available private keys
	// NB: discarding returned error since this is merely a best-effort look-up for convenience
	username, email, _ := s.askForGitConfigUser(ctx, name)
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

func detectCryptoBackend(ctx context.Context, path string) backend.CryptoBackend {
	if fsutil.IsFile(filepath.Join(path, xc.IDFile)) {
		return backend.XC
	}
	return backend.GPGCLI
}
