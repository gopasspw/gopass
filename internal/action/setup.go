package action

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	gpgcli "github.com/gopasspw/gopass/internal/backend/crypto/gpg/cli"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

// Setup will invoke the onboarding / setup wizard.
func (s *Action) Setup(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	remote := c.String("remote")
	team := c.String("alias")
	create := c.Bool("create")

	ctx = initParseContext(ctx, c)

	out.Printf(ctx, logo)
	out.Printf(ctx, "🌟 Welcome to gopass!")
	out.Printf(ctx, "🌟 Initializing a new password store ...")

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
	ctx = age.WithOnlyNative(ctx, true)
	// gpg: only trusted keys
	// only list "usable" / properly trused and signed GPG keys by requesting
	// always trust is false. Ignored for other backends. See
	// https://www.gnupg.org/gph/en/manual/r1554.html.
	ctx = gpg.WithAlwaysTrust(ctx, false)

	// need to re-initialize the root store or it's already initialized
	// and won't properly set up crypto according to our context.
	s.Store = root.New(s.cfg)
	inited, err := s.Store.IsInitialized(ctx)
	if err != nil {
		return exit.Error(exit.Unknown, err, "Failed to check store status: %s", err)
	}

	if inited {
		out.Errorf(ctx, "Store is already initialized. Aborting wizard to avoid overwriting existing data.")

		return nil
	}

	debug.Log("Starting Onboarding Wizard - remote: %s - team: %s - create: %t - name: %s - email: %s", remote, team, create, ctxutil.GetUsername(ctx), ctxutil.GetEmail(ctx))

	crypto := s.getCryptoFor(ctx, team)
	if crypto == nil {
		return fmt.Errorf("can not continue without crypto")
	}
	debug.Log("Crypto Backend initialized as: %s", crypto.Name())

	if err := s.initCheckPrivateKeys(ctx, crypto); err != nil {
		return fmt.Errorf("failed to check private keys: %w", err)
	}

	// if a git remote and a team name are given attempt unattended team setup.
	if remote != "" && team != "" {
		if create {
			return s.initCreateTeam(ctx, team, remote)
		}

		return s.initJoinTeam(ctx, team, remote)
	}

	// assume local setup by default, remotes can be added easily later.
	return s.initLocal(ctx)
}

func (s *Action) initCheckPrivateKeys(ctx context.Context, crypto backend.Crypto) error {
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

	return nil
}

func (s *Action) initGenerateIdentity(ctx context.Context, crypto backend.Crypto, name, email string) error {
	out.Printf(ctx, "🧪 Creating cryptographic key pair (%s) ...", crypto.Name())

	if crypto.Name() == gpgcli.Name {
		var err error

		out.Printf(ctx, "🎩 Gathering information for the %s key pair ...", crypto.Name())
		name, err = termio.AskForString(ctx, "🚶 What is your name?", name)
		if err != nil {
			return err
		}

		email, err = termio.AskForString(ctx, "📧 What is your email?", email)
		if err != nil {
			return err
		}
	}

	passphrase := xkcdgen.Random()
	pwGenerated := true
	want, err := termio.AskForBool(ctx, "⚠ Do you want to enter a passphrase? (otherwise we generate one for you)", false)
	if err != nil {
		return err
	}
	if want {
		pwGenerated = false
		sv, err := termio.AskForPassword(ctx, "passphrase for your new keypair", true)
		if err != nil {
			return fmt.Errorf("failed to read passphrase: %w", err)
		}
		passphrase = sv
	}

	if crypto.Name() == "gpgcli" {
		// Note: This issue shouldn't matter much past Linux Kernel 5.6,
		// eventually we might want to remove this notice. Only applies to
		// GPG.
		out.Printf(ctx, "⏳ This can take a long time. If you get impatient see https://go.gopass.pw/entropy")
		if want, err := termio.AskForBool(ctx, "Continue?", true); err != nil || !want {
			return fmt.Errorf("user aborted: %w", err)
		}
	}

	if err := crypto.GenerateIdentity(ctx, name, email, passphrase); err != nil {
		return fmt.Errorf("failed to create new private key: %w", err)
	}

	out.OKf(ctx, "Key pair generated")

	if pwGenerated {
		out.Printf(ctx, color.MagentaString("Passphrase: ")+passphrase)
		out.Noticef(ctx, "You need to remember this very well!")
	}

	out.Notice(ctx, "🔐 We need to unlock your newly created private key now! Please enter the passphrase you just generated.")

	// avoid the gpg cache or we won't find the newly created key
	kl, err := crypto.ListIdentities(gpg.WithUseCache(ctx, false))
	if err != nil {
		return fmt.Errorf("failed to list private keys: %w", err)
	}

	if len(kl) > 1 {
		out.Notice(ctx, "More than one private key detected. Make sure to use the correct one!")
	}

	if len(kl) < 1 {
		return fmt.Errorf("failed to create a usable key pair")
	}

	// we can export the generated key to the current directory for convenience.
	if err := s.initExportPublicKey(ctx, crypto, kl[0]); err != nil {
		return err
	}
	out.OKf(ctx, "Key pair validated")

	return nil
}

type keyExporter interface {
	ExportPublicKey(ctx context.Context, id string) ([]byte, error)
}

func (s *Action) initExportPublicKey(ctx context.Context, crypto backend.Crypto, key string) error {
	exp, ok := crypto.(keyExporter)
	if !ok {
		debug.Log("crypto backend %T can not export public keys", crypto)

		return nil
	}

	fn := key + ".pub.key"
	want, err := termio.AskForBool(ctx, fmt.Sprintf("Do you want to export your public key to %q?", fn), false)
	if err != nil {
		return err
	}

	if !want {
		return nil
	}

	pk, err := exp.ExportPublicKey(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}

	if err := os.WriteFile(fn, pk, 0o6444); err != nil {
		out.Errorf(ctx, "❌ Failed to export public key %q: %q", fn, err)

		return err
	}
	out.Printf(ctx, "✴ Public key exported to %q", fn)

	return nil
}

func (s *Action) initHasUseablePrivateKeys(ctx context.Context, crypto backend.Crypto) bool {
	debug.Log("checking for existing, usable identities / private keys for %s", crypto.Name())
	kl, err := crypto.ListIdentities(ctx)
	if err != nil {
		return false
	}

	debug.Log("available private keys: %q for %s", kl, crypto.Name())

	return len(kl) > 0
}

func (s *Action) initSetupGitRemote(ctx context.Context, team, remote string) error {
	var err error
	remote, err = termio.AskForString(ctx, "Please enter the git remote for your shared store", remote)
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	// omit RCS output.
	ctx = ctxutil.WithHidden(ctx, true)
	if err := s.Store.RCSAddRemote(ctx, team, "origin", remote); err != nil {
		return fmt.Errorf("failed to add git remote: %w", err)
	}
	// initial pull, in case the remote is non-empty.
	if err := s.Store.RCSPull(ctx, team, "origin", ""); err != nil {
		debug.Log("Initial git pull failed: %s", err)
	}
	if err := s.Store.RCSPush(ctx, team, "origin", ""); err != nil {
		return fmt.Errorf("failed to push to git remote: %w", err)
	}

	return nil
}

// initLocal will initialize a local store, useful for local-only setups or as
// part of team setups to create the root store.
func (s *Action) initLocal(ctx context.Context) error {
	path := ""
	if s.Store != nil {
		path = s.Store.Path()
	}

	out.Printf(ctx, "🌟 Configuring your password store ...")
	if err := s.init(ctxutil.WithHidden(ctx, true), "", path); err != nil {
		return fmt.Errorf("failed to init local store: %w", err)
	}

	if backend.GetStorageBackend(ctx) == backend.GitFS {
		debug.Log("configuring git remotes")
		if want, err := termio.AskForBool(ctx, "❓ Do you want to add a git remote?", false); err == nil && want {
			out.Printf(ctx, "Configuring the git remote ...")
			if err := s.initSetupGitRemote(ctx, "", ""); err != nil {
				return fmt.Errorf("failed to setup git remote: %w", err)
			}
		}
	}
	// TODO remotes for fossil, etc.

	// detect and add mount a for passage
	if err := s.initDetectPassage(ctx); err != nil {
		out.Warningf(ctx, "Failed to add passage mount: %s", err)
	}

	// save config.
	if err := s.cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	out.OKf(ctx, "Configuration written to %s", s.cfg.Path)

	return nil
}

func (s *Action) initDetectPassage(ctx context.Context) error {
	pIds := age.PassageIdFile()
	if !fsutil.IsFile(pIds) {
		debug.Log("no passage identities found at %s", pIds)

		return nil
	}

	pDir := filepath.Dir(pIds)

	if err := s.Store.AddMount(ctx, "passage", pDir); err != nil {
		return fmt.Errorf("failed to mount passage dir: %w", err)
	}

	out.OKf(ctx, "Detected passage store at %s. Mounted below passage/.", pDir)

	return nil
}

// initCreateTeam will create a local root store and a shared team store.
func (s *Action) initCreateTeam(ctx context.Context, team, remote string) error {
	var err error

	out.Printf(ctx, "Creating a new team ...")
	if err := s.initLocal(ctx); err != nil {
		return fmt.Errorf("failed to create local store: %w", err)
	}

	// name of the new team.
	team, err = termio.AskForString(ctx, out.Prefix(ctx)+"Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}
	ctx = out.AddPrefix(ctx, "["+team+"] ")

	out.Printf(ctx, "Initializing your shared store ...")
	if err := s.init(ctxutil.WithHidden(ctx, true), team, ""); err != nil {
		return fmt.Errorf("failed to init shared store: %w", err)
	}
	out.OKf(ctx, "Done. Initialized the store.")

	out.Printf(ctx, "Configuring the git remote ...")
	if err := s.initSetupGitRemote(ctx, team, remote); err != nil {
		return fmt.Errorf("failed to setup git remote: %w", err)
	}
	out.OKf(ctx, "Done. Created Team %q", team)

	return nil
}

// initJoinTeam will create a local root store and clone an existing store to
// a mount.
func (s *Action) initJoinTeam(ctx context.Context, team, remote string) error {
	var err error

	out.Printf(ctx, "Joining existing team ...")
	if err := s.initLocal(ctx); err != nil {
		return fmt.Errorf("failed to create local store: %w", err)
	}

	// name of the existing team.
	team, err = termio.AskForString(ctx, out.Prefix(ctx)+"Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return err
	}
	ctx = out.AddPrefix(ctx, "["+team+"]")

	out.Printf(ctx, "Configuring git remote ...")
	remote, err = termio.AskForString(ctx, out.Prefix(ctx)+"Please enter the git remote for your shared store", remote)
	if err != nil {
		return err
	}

	out.Printf(ctx, "Cloning from the git remote ...")
	if err := s.clone(ctxutil.WithHidden(ctx, true), remote, team, ""); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}
	out.OKf(ctx, "Done. Joined Team %q", team)
	out.Noticef(ctx, "You still need to request access to decrypt secrets!")

	return nil
}
