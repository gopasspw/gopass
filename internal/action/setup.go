package action

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// Setup will invoke the onboarding / setup wizard
func (s *Action) Setup(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	remote := c.String("remote")
	team := c.String("alias")
	create := c.Bool("create")

	ctx = initParseContext(ctx, c)

	out.Print(ctx, logo)
	out.Print(ctx, "🌟 Welcome to gopass!")
	out.Print(ctx, "🌟 Initializing a new password store ...")

	if name := termio.DetectName(c.Context, c); name != "" {
		ctx = ctxutil.WithUsername(ctx, name)
	}
	if email := termio.DetectEmail(c.Context, c); email != "" {
		ctx = ctxutil.WithEmail(ctx, email)
	}
	// need to re-initialize the root store or it's already initialized
	// and won't properly set up crypto according to our context.
	s.Store = root.New(s.cfg)
	inited, err := s.Store.IsInitialized(ctx)
	if err != nil {
		return ExitError(ExitUnknown, err, "Failed to initialized store: %s", err)
	}
	if inited {
		out.Error(ctx, "⚠ Store is already initialized. Aborting.")
		return nil
	}

	debug.Log("Starting Onboarding Wizard - remote: %s - team: %s - create: %t - name: %s - email: %s", remote, team, create, ctxutil.GetUsername(ctx), ctxutil.GetEmail(ctx))

	crypto := s.getCryptoFor(ctx, team)
	if crypto == nil {
		return fmt.Errorf("can not continue without crypto")
	}
	debug.Log("Crypto Backend initialized as: %s", crypto.Name())

	// check for existing GPG/Age keypairs (private/secret keys). We need at least
	// one useable key pair. If none exists try to create one
	if !s.initHasUseablePrivateKeys(ctx, crypto) {
		out.Print(ctx, "🔐 No useable cryptographic keys. Generating new key pair")
		out.Print(ctx, "🕰 Key generation may take up to a few minutes")
		if err := s.initGenerateIdentity(ctx, crypto, ctxutil.GetUsername(ctx), ctxutil.GetEmail(ctx)); err != nil {
			return errors.Wrapf(err, "failed to create new private key")
		}
		out.Print(ctx, "🔐 Cryptographic keys generated")
	}

	debug.Log("We have useable private keys")

	// if a git remote and a team name are given attempt unattended team setup
	if remote != "" && team != "" {
		if create {
			return s.initCreateTeam(ctx, team, remote)
		}
		return s.initJoinTeam(ctx, team, remote)
	}

	// assume local setup by default, remotes can be added easily later
	return s.initLocal(ctx)
}

func (s *Action) initGenerateIdentity(ctx context.Context, crypto backend.Crypto, name, email string) error {
	out.Green(ctx, "🧪 Creating cryptographic key pair (%s) ...", crypto.Name())

	out.Print(ctx, "🎩 Gathering information for the key pair ...")
	name, err := termio.AskForString(ctx, "🚶 What is your name?", name)
	if err != nil {
		return err
	}

	email, err = termio.AskForString(ctx, "📧 What is your email?", email)
	if err != nil {
		return err
	}

	passphrase := xkcdgen.Random()
	pwGenerated := true
	if bv, err := termio.AskForBool(ctx, "⚠ Do you want to enter a passphrase? (otherwise we generate one for you)", false); err != nil && bv {
		pwGenerated = false
		sv, err := termio.AskForPassword(ctx, "✍ Please enter your passphrase")
		if err != nil {
			return errors.Wrapf(err, "Failed to read passphrase")
		}
		passphrase = sv
	}

	// Note: This issue shouldn't matter much past Linux Kernel 5.6,
	// eventually we might want to remove this notice.
	out.Yellow(ctx, "⏳ This can take a long time. If you get impatient see https://github.com/gopasspw/gopass/blob/master/docs/entropy.md")
	if want, err := termio.AskForBool(ctx, "Continue?", true); err != nil || !want {
		return errors.Wrapf(err, "User aborted")
	}

	if err := crypto.GenerateIdentity(ctx, name, email, passphrase); err != nil {
		return errors.Wrapf(err, "failed to create new private key")
	}

	out.Print(ctx, "✅ Key pair generated")

	if pwGenerated {
		out.Print(ctx, color.MagentaString("Passphrase: ")+passphrase)
		out.Print(ctx, "⚠ You need to remember this very well!")
	}

	// avoid the gpg cache or we won't find the newly created key
	kl, err := crypto.ListIdentities(gpg.WithUseCache(ctx, false))
	if err != nil {
		return errors.Wrapf(err, "failed to list private keys")
	}
	if len(kl) > 1 {
		out.Print(ctx, "⚠ More than one private key detected. Make sure to use the correct one!")
		return nil
	}
	if len(kl) < 1 {
		return errors.New("failed to create a usable key pair")
	}

	// we can export the generated key to the current directory for convenience
	if err := s.initExportPublicKey(ctx, crypto, kl[0]); err != nil {
		return err
	}
	out.Green(ctx, "✅ Key pair validated")
	return nil
}

func (s *Action) initExportPublicKey(ctx context.Context, crypto backend.Crypto, key string) error {
	fn := key + ".pub.key"
	want, err := termio.AskForBool(ctx, fmt.Sprintf("Do you want to export your public key to %q?", fn), false)
	if err != nil {
		return err
	}
	if !want {
		return nil
	}
	pk, err := crypto.ExportPublicKey(ctx, key)
	if err != nil {
		return errors.Wrapf(err, "failed to export public key")
	}
	if err := ioutil.WriteFile(fn, pk, 06444); err != nil {
		out.Error(ctx, "❌ Failed to export public key %q: %q", fn, err)
		return err
	}
	out.Print(ctx, "✴ Public key exported to '%s'", fn)
	return nil
}

func (s *Action) initHasUseablePrivateKeys(ctx context.Context, crypto backend.Crypto) bool {
	// only list "usable" / properly trused and signed GPG keys by requesting
	// always trust is false. Ignored for other backends. See
	// https://www.gnupg.org/gph/en/manual/r1554.html
	kl, err := crypto.ListIdentities(gpg.WithAlwaysTrust(ctx, false))
	if err != nil {
		return false
	}
	debug.Log("available private keys: %+v", kl)
	return len(kl) > 0
}

func (s *Action) initSetupGitRemote(ctx context.Context, team, remote string) error {
	var err error
	remote, err = termio.AskForString(ctx, "Please enter the git remote for your shared store", remote)
	if err != nil {
		return errors.Wrapf(err, "failed to read user input")
	}

	// omit RCS output
	ctx = ctxutil.WithHidden(ctx, true)
	if err := s.Store.RCSAddRemote(ctx, team, "origin", remote); err != nil {
		return errors.Wrapf(err, "failed to add git remote")
	}
	// initial pull, in case the remote is non-empty
	if err := s.Store.RCSPull(ctx, team, "origin", "master"); err != nil {
		debug.Log("Initial git pull failed: %s", err)
	}
	if err := s.Store.RCSPush(ctx, team, "origin", "master"); err != nil {
		return errors.Wrapf(err, "failed to push to git remote")
	}
	return nil
}

// initLocal will initialize a local store, useful for local-only setups or as
// part of team setups to create the root store
func (s *Action) initLocal(ctx context.Context) error {
	path := ""
	if s.Store != nil {
		path = s.Store.Path()
	}

	out.Print(ctx, "🌟 Configuring your password store ...")
	if err := s.init(ctxutil.WithHidden(ctx, true), "", path); err != nil {
		return errors.Wrapf(err, "failed to init local store")
	}

	if want, err := termio.AskForBool(ctx, "❓ Do you want to add a git remote?", false); err == nil && want {
		out.Print(ctx, "Configuring the git remote ...")
		if err := s.initSetupGitRemote(ctx, "", ""); err != nil {
			return errors.Wrapf(err, "failed to setup git remote")
		}
	}

	// save config
	if err := s.cfg.Save(); err != nil {
		return errors.Wrapf(err, "failed to save config")
	}

	out.Green(ctx, "✅ Configured")
	return nil
}

// initCreateTeam will create a local root store and a shared team store
func (s *Action) initCreateTeam(ctx context.Context, team, remote string) error {
	var err error

	out.Print(ctx, "Creating a new team ...")
	if err := s.initLocal(ctx); err != nil {
		return errors.Wrapf(err, "failed to create local store")
	}

	// name of the new team
	team, err = termio.AskForString(ctx, out.Prefix(ctx)+"Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return errors.Wrapf(err, "failed to read user input")
	}
	ctx = out.AddPrefix(ctx, "["+team+"] ")

	out.Print(ctx, "Initializing your shared store ...")
	if err := s.init(ctxutil.WithHidden(ctx, true), team, ""); err != nil {
		return errors.Wrapf(err, "failed to init shared store")
	}
	out.Print(ctx, "✅ Done. Initialized the store.")

	out.Print(ctx, "Configuring the git remote ...")
	if err := s.initSetupGitRemote(ctx, team, remote); err != nil {
		return errors.Wrapf(err, "failed to setup git remote")
	}
	out.Print(ctx, "✅ Done. Created Team %q", team)
	return nil
}

// initJoinTeam will create a local root store and clone an existing store to
// a mount
func (s *Action) initJoinTeam(ctx context.Context, team, remote string) error {
	var err error

	out.Print(ctx, "Joining existing team ...")
	if err := s.initLocal(ctx); err != nil {
		return errors.Wrapf(err, "failed to create local store")
	}

	// name of the existing team
	team, err = termio.AskForString(ctx, out.Prefix(ctx)+"Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return err
	}
	ctx = out.AddPrefix(ctx, "["+team+"]")

	out.Print(ctx, "Configuring git remote ...")
	remote, err = termio.AskForString(ctx, out.Prefix(ctx)+"Please enter the git remote for your shared store", remote)
	if err != nil {
		return err
	}

	out.Print(ctx, "Cloning from the git remote ...")
	if err := s.clone(ctxutil.WithHidden(ctx, true), remote, team, ""); err != nil {
		return errors.Wrapf(err, "failed to clone repo")
	}
	out.Print(ctx, "✅ Done. Joined Team %q", team)
	out.Print(ctx, "⚠ You still need to request access to decrypt secrets!")
	return nil
}
