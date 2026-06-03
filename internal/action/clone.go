package action

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v3"
)

// Clone will fetch and mount a new password store from a git repo.
// It can also be used to clone a new password store to a submount.
func (s *setupHandler) Clone(ctx context.Context, cmd *cli.Command) error {
	ctx = ctxutil.WithGlobalFlags(ctx, cmd)
	if cmd.IsSet("crypto") {
		var err error
		ctx, err = backend.WithCryptoBackendString(ctx, cmd.String("crypto"))
		if err != nil {
			return exit.Error(exit.Unknown, err, "Failed to set crypto backend: %s", err)
		}
	}

	if cmd.IsSet("storage") {
		var err error
		ctx, err = backend.WithStorageBackendString(ctx, cmd.String("storage"))
		if err != nil {
			return exit.Error(exit.Unknown, err, "Failed to set storage backend: %s", err)
		}
	}

	path := cmd.String("path")

	if cmd.Args().Len() < 1 {
		return exit.Error(exit.Usage, nil, "Usage: %s clone repo [mount]", s.Name)
	}

	// gopass clone [--crypto=foo] [--path=/some/store] git://foo/bar team0.
	repo := cmd.Args().Get(0)
	mount := ""
	if cmd.Args().Len() > 1 {
		mount = cmd.Args().Get(1)
	}

	out.Printf(ctx, logo)
	out.Printf(ctx, "🌟 Welcome to gopass!")
	out.Printf(ctx, "🌟 Cloning an existing password store from %q ...", repo)

	if name := termio.DetectName(ctx, cmd); name != "" {
		ctx = ctxutil.WithUsername(ctx, name)
	}
	if email := termio.DetectEmail(ctx, cmd); email != "" {
		ctx = ctxutil.WithEmail(ctx, email)
	}

	// age: only native keys
	// "[ssh] types should only be used for compatibility with existing keys,
	// and native X25519 keys should be preferred otherwise."
	// https://pkg.go.dev/filippo.io/age@v1.0.0/agessh#pkg-overview.
	ctx = age.WithOnlyNative(ctx, true)
	// gpg: only trusted keys
	// only list "usable" / properly trused and signed GPG keys by requesting
	// always trust is false. Ignored for other backends. See
	// https://www.gnupg.org/gph/en/manual/r1554.html.
	ctx = gpg.WithAlwaysTrust(ctx, false)

	if err := s.clone(ctx, repo, mount, path); err != nil {
		return err
	}

	// need to re-initialize the root store or it's already initialized
	// and won't properly set up crypto according to our context.
	s.Store = root.New(s.cfg)
	inited, err := s.Store.IsInitialized(ctx)
	if err != nil {
		return exit.Error(exit.Unknown, err, "Failed to check store status: %s", err)
	}

	if !inited {
		out.Errorf(ctx, "Failed to clone")

		return nil
	}

	if !cmd.Bool("check-keys") {
		return nil
	}

	// Unified join flow (Stage 2 / GH-2620): imports existing .public-keys/,
	// checks decryption, and — if needed — exports the user's own key additively.
	return s.cloneJoinTeam(ctx, mount)
}

// cloneJoinTeam performs the unified post-clone join processing: import
// existing keys, check decryption, and if needed export the user's key
// additively (never removing other recipients). This replaces the old
// cloneCheckDecryptionKeys path which could regenerate a reduced key set.
func (s *setupHandler) cloneJoinTeam(ctx context.Context, mount string) error {
	exported, err := s.Store.JoinTeam(ctx, mount)
	if err != nil {
		out.Warningf(ctx, "Join team processing: %s", err)

		return nil
	}

	if exported {
		out.Noticef(ctx, "🔑 Your public key has been added to the store's .public-keys/.")
		out.Noticef(ctx, "Request access: ask a team owner to run 'gopass recipients add <your-key>' and 'gopass sync'.")
	} else {
		out.OKf(ctx, "You can decrypt this store. Welcome to the team!")
	}

	return nil
}

// storageBackendOrDefault will return a storage backend that can be clone,
// i.e. specifically backend.FS can't be cloned.
func storageBackendOrDefault(ctx context.Context, repo string) backend.StorageBackend {
	// first try to get it from the context.
	if be := backend.GetStorageBackend(ctx); be != backend.FS {
		return be
	}

	if strings.HasSuffix(repo, ".fossil") {
		return backend.FossilFS
	}

	if strings.HasSuffix(repo, ".git") {
		return backend.GitFS
	}

	debug.Log("falling back to the default storage backend for clone (GitFS)")

	return backend.GitFS
}

func (s *setupHandler) clone(ctx context.Context, repo, mount, path string) error {
	if path == "" {
		path = config.PwStoreDir(mount)
	}

	inited, err := s.Store.IsInitialized(ctxutil.WithGitInit(ctx, false))
	if err != nil {
		return exit.Error(exit.Unknown, err, "Failed to initialized stores: %s", err)
	}

	if mount == "" && inited {
		return exit.Error(exit.AlreadyInitialized, nil, "Cannot clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
	}

	// make sure the parent directory exists.
	if parentPath := filepath.Dir(path); !fsutil.IsDir(parentPath) {
		if err := os.MkdirAll(parentPath, 0o700); err != nil {
			return exit.Error(exit.Unknown, err, "Failed to create parent directory for clone: %s", err)
		}
	}

	// clone repo.
	sb := storageBackendOrDefault(ctx, repo)
	out.Noticef(ctx, "Cloning %s repository %q to %q ...", sb, repo, path)
	_, err = backend.Clone(ctx, sb, repo, path)
	if err != nil {
		return exit.Error(exit.Git, err, "failed to clone repo %q to %q: %s", repo, path, err)
	}

	// add mount.
	debug.Log("Mounting cloned repo %q at %q", path, mount)
	if err := s.cloneAddMount(ctx, mount, path); err != nil {
		return err
	}

	// try to init repo config.
	out.Noticef(ctx, "Configuring %s repository ...", sb)

	// ask for config values.
	username, email, err := s.cloneGetGitConfig(ctx, mount)
	if err != nil {
		return err
	}

	// initialize repo config.
	if err := s.Store.RCSInitConfig(ctx, mount, username, email); err != nil {
		debug.Log("Stacktrace: %+v\n", err)
		out.Errorf(ctx, "Failed to configure %s: %s", sb, err)
	}

	if mount != "" {
		mount = " " + mount
	}

	out.Printf(ctx, "Your password store is ready to use! Have a look around: `%s list%s`\n", s.Name, mount)

	return nil
}

func (s *setupHandler) cloneAddMount(ctx context.Context, mount, path string) error {
	if mount == "" {
		return nil
	}

	inited, err := s.Store.IsInitialized(ctx)
	if err != nil {
		return exit.Error(exit.Unknown, err, "Failed to initialize store: %s", err)
	}

	if !inited {
		return exit.Error(exit.NotInitialized, nil, "Root-Store is not initialized. Clone or init root store first")
	}

	if err := s.Store.AddMount(ctx, mount, path); err != nil {
		return exit.Error(exit.Mount, err, "Failed to add mount: %s", err)
	}
	out.Printf(ctx, "Mounted password store %s at mount point `%s` ...", path, mount)

	return nil
}

func (s *setupHandler) cloneGetGitConfig(ctx context.Context, name string) (string, string, error) {
	out.Printf(ctx, "🎩 Gathering information for the git repository ...")
	// for convenience, set defaults to user-selected values from available private keys.
	// NB: discarding returned error since this is merely a best-effort look-up for convenience.
	username, email, _ := cui.AskForGitConfigUser(ctx, s.Store.Crypto(ctx, name))
	if username == "" {
		username = termio.DetectName(ctx, nil)
		var err error
		username, err = termio.AskForString(ctx, "🚶 What is your name?", username)
		if err != nil {
			return "", "", exit.Error(exit.IO, err, "Failed to read user input: %s", err)
		}
	}

	if email == "" {
		email = termio.DetectEmail(ctx, nil)
		var err error
		email, err = termio.AskForString(ctx, "📧 What is your email?", email)
		if err != nil {
			return "", "", exit.Error(exit.IO, err, "Failed to read user input: %s", err)
		}
	}

	return username, email, nil
}
