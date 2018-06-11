package action

import (
	"context"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/backend/crypto/xc"
	gitcli "github.com/gopasspw/gopass/pkg/backend/rcs/git/cli"
	"github.com/gopasspw/gopass/pkg/backend/rcs/git/gogit"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/cui"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Clone will fetch and mount a new password store from a git repo
func (s *Action) Clone(ctx context.Context, c *cli.Context) error {
	if c.IsSet("crypto") {
		ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))
	}
	if c.IsSet("sync") {
		ctx = backend.WithRCSBackendString(ctx, c.String("sync"))
	}

	if len(c.Args()) < 1 {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s clone repo [mount]", s.Name)
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
	inited, err := s.Store.Initialized(ctx)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "Failed to initialized stores: %s", err)
	}
	if mount == "" && inited {
		return ExitError(ctx, ExitAlreadyInitialized, nil, "Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
	}

	// clone repo
	switch backend.GetRCSBackend(ctx) {
	case backend.GoGit:
		if _, err := gogit.Clone(ctx, repo, path); err != nil {
			return ExitError(ctx, ExitGit, err, "failed to clone repo '%s' to '%s'", repo, path)
		}
	case backend.GitCLI:
		fallthrough
	default:
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
		if _, err := gitcli.Clone(ctx, repo, path); err != nil {
			return ExitError(ctx, ExitGit, err, "failed to clone repo '%s' to '%s'", repo, path)
		}
	}

	// detect crypto backend based on cloned repo
	ctx = backend.WithCryptoBackend(ctx, detectCryptoBackend(ctx, path))

	// add mount
	if err := s.cloneAddMount(ctx, mount, path); err != nil {
		return err
	}

	// save new mount in config file
	if err := s.cfg.Save(); err != nil {
		return ExitError(ctx, ExitIO, err, "Failed to update config: %s", err)
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

func (s *Action) cloneAddMount(ctx context.Context, mount, path string) error {
	if mount == "" {
		s.cfg.Root.Path.Crypto = backend.GetCryptoBackend(ctx)
		s.cfg.Root.Path.RCS = backend.GetRCSBackend(ctx)
		s.cfg.Root.Path.Storage = backend.GetStorageBackend(ctx)
		return nil
	}

	inited, err := s.Store.Initialized(ctx)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "Failed to initialize store: %s", err)
	}
	if !inited {
		return ExitError(ctx, ExitNotInitialized, nil, "Root-Store is not initialized. Clone or init root store first")
	}
	if err := s.Store.AddMount(ctx, mount, path); err != nil {
		return ExitError(ctx, ExitMount, err, "Failed to add mount: %s", err)
	}
	out.Green(ctx, "Mounted password store %s at mount point `%s` ...", path, mount)
	s.cfg.Mounts[mount].Path.Crypto = backend.GetCryptoBackend(ctx)
	s.cfg.Mounts[mount].Path.RCS = backend.GetRCSBackend(ctx)
	s.cfg.Mounts[mount].Path.Storage = backend.GetStorageBackend(ctx)
	return nil
}

func (s *Action) cloneGetGitConfig(ctx context.Context, name string) (string, string, error) {
	// for convenience, set defaults to user-selected values from available private keys
	// NB: discarding returned error since this is merely a best-effort look-up for convenience
	username, email, _ := cui.AskForGitConfigUser(ctx, s.Store.Crypto(ctx, name), name)
	if username == "" {
		var err error
		username, err = termio.AskForString(ctx, color.CyanString("Please enter a user name for password store git config"), username)
		if err != nil {
			return "", "", ExitError(ctx, ExitIO, err, "Failed to read user input: %s", err)
		}
	}
	if email == "" {
		var err error
		email, err = termio.AskForString(ctx, color.CyanString("Please enter an email address for password store git config"), email)
		if err != nil {
			return "", "", ExitError(ctx, ExitIO, err, "Failed to read user input: %s", err)
		}
	}
	return username, email, nil
}

// detectCryptoBackend tries to detect the crypto backend used in a cloned repo
// This detection is very shallow and doesn't support all backends, yet
func detectCryptoBackend(ctx context.Context, path string) backend.CryptoBackend {
	if fsutil.IsFile(filepath.Join(path, xc.IDFile)) {
		return backend.XC
	}
	return backend.GPGCLI
}
