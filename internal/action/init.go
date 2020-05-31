package action

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/urfave/cli/v2"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// Initialized returns an error if the store is not properly
// prepared.
func (s *Action) Initialized(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	inited, err := s.Store.Initialized(ctx)
	if err != nil {
		return ExitError(ExitUnknown, err, "Failed to initialize store: %s", err)
	}
	if inited {
		debug.Log("Store is already initialized")
		return nil
	}

	debug.Log("Store needs to be initialized")
	if !ctxutil.IsInteractive(ctx) {
		return ExitError(ExitNotInitialized, nil, "password-store is not initialized. Try '%s init'", s.Name)
	}
	if ok, err := termio.AskForBool(ctx, "It seems you are new to gopass. Do you want to run the onboarding wizard?", true); err == nil && ok {
		c.Context = ctx
		if err := s.InitOnboarding(c); err != nil {
			return ExitError(ExitUnknown, err, "failed to run onboarding wizard: %s", err)
		}
		return nil
	}
	return nil
}

// Init a new password store with a first gpg id
func (s *Action) Init(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	path := c.String("path")
	alias := c.String("store")

	ctx = initParseContext(ctx, c)
	if name := termio.DetectName(c.Context, c); name != "" {
		ctx = ctxutil.WithUsername(ctx, name)
	}
	if email := termio.DetectEmail(c.Context, c); email != "" {
		ctx = ctxutil.WithEmail(ctx, email)
	}
	inited, err := s.Store.Initialized(ctx)
	if err != nil {
		return ExitError(ExitUnknown, err, "Failed to initialized store: %s", err)
	}
	if inited {
		out.Error(ctx, "WARNING: Store is already initialized")
	}

	if err := s.init(ctx, alias, path, c.Args().Slice()...); err != nil {
		return ExitError(ExitUnknown, err, "failed to initialize store: %s", err)
	}
	return nil
}

func initParseContext(ctx context.Context, c *cli.Context) context.Context {
	if c.IsSet("crypto") {
		ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))
	}
	if c.IsSet("rcs") {
		ctx = backend.WithRCSBackendString(ctx, c.String("rcs"))
	}
	if c.IsSet("storage") {
		ctx = backend.WithStorageBackendString(ctx, c.String("storage"))
	}

	if !backend.HasCryptoBackend(ctx) {
		debug.Log("Using default Crypto Backend (GPGCLI)")
		ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)
	}
	if !backend.HasRCSBackend(ctx) {
		debug.Log("Using default RCS backend (GitCLI)")
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	}
	if !backend.HasStorageBackend(ctx) {
		debug.Log("Using default storage backend (FS)")
		ctx = backend.WithStorageBackend(ctx, backend.FS)
	}

	ctx = out.WithPrefix(ctx, "[init] ")
	out.Cyan(ctx, "Initializing a new password store ...")

	return ctx
}

func (s *Action) init(ctx context.Context, alias, path string, keys ...string) error {
	if path == "" {
		if alias != "" {
			path = config.PwStoreDir(alias)
		} else {
			path = s.Store.Path()
		}
	}
	debug.Log("action.init(%s, %s, %+v)", alias, path, keys)

	debug.Log("Checking private keys ...")
	crypto := s.getCryptoFor(ctx, alias)
	// private key selection doesn't matter for plain. save one question.
	if crypto.Name() == "plain" {
		keys, _ = crypto.ListIdentities(ctx)
	}
	if len(keys) < 1 {
		nk, err := cui.AskForPrivateKey(ctx, crypto, color.CyanString("Please select a private key for encrypting secrets:"))
		if err != nil {
			return errors.Wrapf(err, "failed to read user input")
		}
		keys = []string{nk}
	}

	debug.Log("Initializing sub store - Alias: %s - Path: %s - Keys: %+v", alias, path, keys)
	if err := s.Store.Init(ctx, alias, path, keys...); err != nil {
		return errors.Wrapf(err, "failed to init store '%s' at '%s'", alias, path)
	}

	if alias != "" && path != "" {
		debug.Log("Mounting sub store %s -> %s", alias, path)
		if err := s.Store.AddMount(ctx, alias, path); err != nil {
			return errors.Wrapf(err, "failed to add mount '%s'", alias)
		}
	}

	if backend.HasRCSBackend(ctx) {
		bn := backend.RCSBackendName(backend.GetRCSBackend(ctx))
		debug.Log("Initializing RCS (%s) ...", bn)
		if err := s.rcsInit(ctx, alias, ctxutil.GetUsername(ctx), ctxutil.GetEmail(ctx)); err != nil {
			debug.Log("Stacktrace: %+v\n", err)
			out.Error(ctx, "Failed to init RCS (%s): %s", bn, err)
		}
	} else {
		debug.Log("not initializing RCS backend ...")
	}

	out.Green(ctx, "Password store %s initialized for:", path)
	s.printRecipients(ctx, alias)

	// write config
	if err := s.cfg.Save(); err != nil {
		return ExitError(ExitConfig, err, "failed to write config: %s", err)
	}

	return nil
}

func (s *Action) printRecipients(ctx context.Context, alias string) {
	crypto := s.Store.Crypto(ctx, alias)
	for _, recipient := range s.Store.ListRecipients(ctx, alias) {
		r := "0x" + recipient
		if kl, err := crypto.FindRecipients(ctx, recipient); err == nil && len(kl) > 0 {
			r = crypto.FormatKey(ctx, kl[0], "")
		}
		out.Yellow(ctx, "  "+r)
	}
}

func (s *Action) getCryptoFor(ctx context.Context, name string) backend.Crypto {
	return s.Store.Crypto(ctx, name)
}

// InitOnboarding will invoke the onboarding / setup wizard
func (s *Action) InitOnboarding(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	remote := c.String("remote")
	team := c.String("alias")
	create := c.Bool("create")
	name := termio.DetectName(c.Context, c)
	if name != "" {
		ctx = ctxutil.WithUsername(ctx, name)
	}
	email := termio.DetectEmail(c.Context, c)
	if email != "" {
		ctx = ctxutil.WithEmail(ctx, email)
	}
	ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))

	// default to git
	if rcs := c.String("rcs"); rcs != "" {
		ctx = backend.WithRCSBackendString(ctx, c.String("rcs"))
	} else {
		debug.Log("Using default RCS backend (GitCLI)")
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	}

	ctx = out.AddPrefix(ctx, "[init] ")
	debug.Log("Starting Onboarding Wizard - remote: %s - team: %s - create: %t - name: %s - email: %s", remote, team, create, name, email)

	crypto := s.getCryptoFor(ctx, name)

	debug.Log("Crypto Backend initialized as: %s", crypto.Name())

	// check for existing GPG keypairs (private/secret keys). We need at least
	// one useable key pair. If none exists try to create one
	if !s.initHasUseablePrivateKeys(ctx, crypto) {
		out.Yellow(ctx, "No useable crypto keys. Generating new key pair")
		ctx := out.AddPrefix(ctx, "[crypto] ")
		out.Print(ctx, "Key generation may take up to a few minutes")
		if err := s.initGenerateIdentity(ctx, crypto, name, email); err != nil {
			return errors.Wrapf(err, "failed to create new private key")
		}
	}

	debug.Log("Has useable private keys")

	// if a git remote and a team name are given attempt unattended team setup
	if remote != "" && team != "" {
		if create {
			return s.initCreateTeam(ctx, team, remote)
		}
		return s.initJoinTeam(ctx, team, remote)
	}

	// no flags given, run interactively
	choices := []string{
		"Local store",
		"Create a Team",
		"Join an existing Team",
	}
	act, sel := cui.GetSelection(ctx, "Select action", choices)
	switch act {
	case "default":
		fallthrough
	case "show":
		switch sel {
		case 0:
			return s.initLocal(ctx)
		case 1:
			return s.initCreateTeam(ctx, "", "")
		case 2:
			return s.initJoinTeam(ctx, "", "")
		}
	default:
		return fmt.Errorf("user aborted")
	}
	return nil
}

func (s *Action) initGenerateIdentity(ctx context.Context, crypto backend.Crypto, name, email string) error {
	out.Green(ctx, "Creating key pair ...")
	out.Yellow(ctx, "WARNING: We are about to generate some GPG keys.")
	out.Print(ctx, `However, the GPG program can sometimes lock up, displaying the following:
"We need to generate a lot of random bytes."
If this happens, please see the following tips:
https://github.com/gopasspw/gopass/blob/master/docs/entropy.md`)
	name, err := termio.AskForString(ctx, "What is your name?", name)
	if err != nil {
		return err
	}

	email, err = termio.AskForString(ctx, "What is your email?", email)
	if err != nil {
		return err
	}

	if want, err := termio.AskForBool(ctx, "Continue?", true); err != nil || !want {
		return errors.Wrapf(err, "User aborted")
	}
	passphrase := xkcdgen.Random()
	if err := crypto.GenerateIdentity(ctx, name, email, passphrase); err != nil {
		return errors.Wrapf(err, "failed to create new private key in batch mode")
	}
	out.Green(ctx, "-> OK")
	out.Print(ctx, color.MagentaString("Passphrase: ")+color.HiGreenString(passphrase))

	kl, err := crypto.ListIdentities(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to list private keys")
	}
	if len(kl) > 1 {
		out.Cyan(ctx, "WARNING: More than one private key detected. Make sure to communicate the right one")
		return nil
	}
	if len(kl) < 1 {
		debug.Log("Private Keys: %+v", kl)
		return errors.New("failed to create a useable key pair")
	}
	key := kl[0]
	fn := key + ".pub.key"
	pk, err := crypto.ExportPublicKey(ctx, key)
	if err != nil {
		return errors.Wrapf(err, "failed to export public key")
	}
	_ = ioutil.WriteFile(fn, pk, 06444)
	out.Cyan(ctx, "Public key exported to '%s'", fn)
	out.Green(ctx, "Done")
	return nil
}

func (s *Action) initHasUseablePrivateKeys(ctx context.Context, crypto backend.Crypto) bool {
	kl, err := crypto.ListIdentities(gpg.WithAlwaysTrust(ctx, false))
	if err != nil {
		return false
	}
	return len(kl) > 0
}

func (s *Action) initSetupGitRemote(ctx context.Context, team, remote string) error {
	var err error
	remote, err = termio.AskForString(ctx, "Please enter the git remote for your shared store", remote)
	if err != nil {
		return errors.Wrapf(err, "failed to read user input")
	}
	{
		ctx := out.WithHidden(ctx, true)
		if err := s.Store.GitAddRemote(ctx, team, "origin", remote); err != nil {
			return errors.Wrapf(err, "failed to add git remote")
		}
		// initial pull, in case the remote is non-empty
		if err := s.Store.GitPull(ctx, team, "origin", "master"); err != nil {
			debug.Log("Initial git pull failed: %s", err)
		}
		if err := s.Store.GitPush(ctx, team, "origin", "master"); err != nil {
			return errors.Wrapf(err, "failed to push to git remote")
		}
	}
	return nil
}

// initLocal will initialize a local store, useful for local-only setups or as
// part of team setups to create the root store
func (s *Action) initLocal(ctx context.Context) error {
	ctx = out.AddPrefix(ctx, "[local] ")

	path := ""
	if s.Store != nil {
		path = s.Store.Path()
	}

	out.Print(ctx, "Initializing your local store ...")
	if err := s.init(out.WithHidden(ctx, true), "", path); err != nil {
		return errors.Wrapf(err, "failed to init local store")
	}
	out.Green(ctx, " -> OK")

	out.Print(ctx, "Configuring your local store ...")

	if want, err := termio.AskForBool(ctx, out.Prefix(ctx)+"Do you want to add a git remote?", false); err == nil && want {
		out.Print(ctx, "Configuring the git remote ...")
		if err := s.initSetupGitRemote(ctx, "", ""); err != nil {
			return errors.Wrapf(err, "failed to setup git remote")
		}
	}

	// noconfirm
	if want, err := termio.AskForBool(ctx, out.Prefix(ctx)+"Do you want to always confirm recipients when encrypting?", true); err == nil {
		s.cfg.ConfirmRecipients = !want
	}

	// save config
	if err := s.cfg.Save(); err != nil {
		return errors.Wrapf(err, "failed to save config")
	}

	out.Green(ctx, " -> OK")
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
	if err := s.init(out.WithHidden(ctx, true), team, ""); err != nil {
		return errors.Wrapf(err, "failed to init shared store")
	}
	out.Green(ctx, " -> OK")

	out.Print(ctx, "Configuring the git remote ...")
	if err := s.initSetupGitRemote(ctx, team, remote); err != nil {
		return errors.Wrapf(err, "failed to setup git remote")
	}
	out.Green(ctx, " -> OK")
	out.Green(ctx, "Created Team '%s'", team)
	return nil
}

// initJoinTeam will create a local root store and clone and existing store to
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
	if err := s.clone(out.WithHidden(ctx, true), remote, team, ""); err != nil {
		return errors.Wrapf(err, "failed to clone repo")
	}
	out.Green(ctx, " -> OK")
	out.Green(ctx, "Joined Team '%s'", team)
	out.Yellow(ctx, "Note: You still need to request access to decrypt any secret!")
	return nil
}
