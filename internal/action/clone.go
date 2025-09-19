package action

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend"
	agecrypto "github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/set"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Clone will fetch and mount a new password store from a git repo.
func (s *Action) Clone(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("crypto") {
		ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))
	}

	if c.IsSet("storage") {
		ctx = backend.WithStorageBackendString(ctx, c.String("storage"))
	}

	path := c.String("path")

	if c.Args().Len() < 1 {
		return exit.Error(exit.Usage, nil, "Usage: %s clone repo [mount]", s.Name)
	}

	// gopass clone [--crypto=foo] [--path=/some/store] git://foo/bar team0.
	repo := c.Args().Get(0)
	mount := ""
	if c.Args().Len() > 1 {
		mount = c.Args().Get(1)
	}

	out.Printf(ctx, logo)
	out.Printf(ctx, "🌟 Welcome to gopass!")
	out.Printf(ctx, "🌟 Cloning an existing password store from %q ...", repo)

	if name := termio.DetectName(ctx, c); name != "" {
		ctx = ctxutil.WithUsername(ctx, name)
	}
	if email := termio.DetectEmail(ctx, c); email != "" {
		ctx = ctxutil.WithEmail(ctx, email)
	}

	// age: only native keys
	// "[ssh] types should only be used for compatibility with existing keys,
	// and native X25519 keys should be preferred otherwise."
	// https://pkg.go.dev/filippo.io/age@v1.0.0/agessh#pkg-overview.
	ctx = agecrypto.WithOnlyNative(ctx, true)
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

	if !c.Bool("check-keys") {
		return nil
	}

	return s.cloneCheckDecryptionKeys(ctx, mount)
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

func (s *Action) clone(ctx context.Context, repo, mount, path string) error {
	if path == "" {
		path = config.PwStoreDir(mount)
	}

	inited, err := s.Store.IsInitialized(ctxutil.WithGitInit(ctx, false))
	if err != nil {
		return exit.Error(exit.Unknown, err, "Failed to initialized stores: %s", err)
	}

	if mount == "" && inited {
		return exit.Error(exit.AlreadyInitialized, nil, "Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
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
	if _, err := backend.Clone(ctx, sb, repo, path); err != nil {
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

func (s *Action) cloneCheckDecryptionKeys(ctx context.Context, mount string) error {
	crypto := s.getCryptoFor(ctx, mount)
	if crypto == nil {
		return fmt.Errorf("can not continue without crypto")
	}
	debug.Log("Crypto Backend initialized as: %s", crypto.Name())

	// check for existing GPG/Age keypairs (private/secret keys). We need at least
	// one useable key pair. If none exists try to create one.
	if !s.initHasUseablePrivateKeys(ctx, crypto) {
		out.Printf(ctx, "🔐 No useable cryptographic keys. Generating new key pair")
		if crypto.Name() == "gpgcli" {
			out.Printf(ctx, "🕰 Key generation may take up to a few minutes")
		}
		if err := s.initGenerateIdentity(ctx, crypto, ctxutil.GetUsername(ctx), ctxutil.GetEmail(ctx)); err != nil {
			return fmt.Errorf("failed to create new private key: %w", err)
		}
		out.Printf(ctx, "🔐 Cryptographic keys generated")
	}

	debug.Log("We have useable private keys")

	recpSet := set.New(s.Store.ListRecipients(ctx, mount)...)
	ids, err := crypto.ListIdentities(ctx)
	if err != nil {
		out.Warningf(ctx, "Failed to check decryption keys: %s", err)

		return nil
	}

	idSet := set.New(ids...)
	// Check whether any of our usable keys are in recpSet
	if _, found := recpSet.Choose(idSet.Contains); found {
		out.Noticef(ctx, "Found valid decryption keys. You can now decrypt your passwords.")

		return nil
	}

	var exported bool
	if sub, err := s.Store.GetSubStore(mount); err == nil {
		debug.Log("exporting public keys: %v", idSet.Elements())
		exported, err = sub.UpdateExportedPublicKeys(ctx)
		if err != nil {
			debug.Log("failed to export missing public keys: %w", err)
		}
	} else {
		debug.Log("failed to get sub store: %s", err)
	}

	out.Noticef(ctx, "Please ask the owner of the password store to add one of your keys: %s", strings.Join(idSet.Elements(), ", "))
	if exported {
		out.Noticef(ctx, "The missing keys were exported to the password store. Run `gopass sync` to push them.")
	}

	return nil
}

func (s *Action) cloneAddMount(ctx context.Context, mount, path string) error {
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

func (s *Action) cloneGetGitConfig(ctx context.Context, name string) (string, string, error) {
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
