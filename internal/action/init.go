package action

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/cui"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/gopasspw/gopass/pkg/store/sub"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// Initialized returns an error if the store is not properly
// prepared.
func (s *Action) Initialized(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	inited, err := s.Store.Initialized(ctx)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "Failed to initialize store: %s", err)
	}
	if inited {
		out.Debug(ctx, "Store is already initialized")
		return nil
	}

	out.Debug(ctx, "Store needs to be initialized")
	if !ctxutil.IsInteractive(ctx) {
		return ExitError(ctx, ExitNotInitialized, nil, "password-store is not initialized. Try '%s init'", s.Name)
	}
	if ok, err := termio.AskForBool(ctx, "It seems you are new to gopass. Do you want to run the onboarding wizard?", true); err == nil && ok {
		c.Context = ctx
		if err := s.InitOnboarding(c); err != nil {
			return ExitError(ctx, ExitUnknown, err, "failed to run onboarding wizard: %s", err)
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
	if name := detectName(c); name != "" {
		ctx = WithUsername(ctx, name)
	}
	if email := detectEmail(c); email != "" {
		ctx = WithEmail(ctx, email)
	}
	inited, err := s.Store.Initialized(ctx)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "Failed to initialized store: %s", err)
	}
	if inited {
		out.Error(ctx, "WARNING: Store is already initialized")
	}

	if err := s.init(ctx, alias, path, c.Args().Slice()...); err != nil {
		return ExitError(ctx, ExitUnknown, err, "failed to initialize store: %s", err)
	}
	return nil
}

func initParseContext(ctx context.Context, c *cli.Context) context.Context {
	if c.IsSet("crypto") {
		ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))
	}
	if c.IsSet("rcs") {
		ctx = backend.WithRCSBackendString(ctx, c.String("rcs"))
	} else {
		if c.IsSet("nogit") && c.Bool("nogit") {
			out.Error(ctx, "DEPRECATION WARNING: Use '--rcs noop' instead")
			ctx = backend.WithRCSBackend(ctx, backend.Noop)
		}
	}

	// default to git
	if !backend.HasRCSBackend(ctx) {
		out.Debug(ctx, "Using default RCS backend (GitCLI)")
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
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
	out.Debug(ctx, "action.init(%s, %s, %+v)", alias, path, keys)

	out.Debug(ctx, "Checking private keys ...")
	crypto := s.getCryptoFor(ctx, alias)
	// private key selection doesn't matter for plain. save one question.
	if crypto.Name() == "plain" {
		keys, _ = crypto.ListPrivateKeyIDs(ctx)
	}
	if len(keys) < 1 {
		nk, err := cui.AskForPrivateKey(ctx, crypto, alias, color.CyanString("Please select a private key for encrypting secrets:"))
		if err != nil {
			return errors.Wrapf(err, "failed to read user input")
		}
		keys = []string{nk}
	}

	out.Debug(ctx, "Initializing sub store - Alias: %s - Path: %s - Keys: %+v", alias, path, keys)
	if err := s.Store.Init(ctx, alias, path, keys...); err != nil {
		return errors.Wrapf(err, "failed to init store '%s' at '%s'", alias, path)
	}

	if alias != "" && path != "" {
		out.Debug(ctx, "Mounting sub store %s -> %s", alias, path)
		if err := s.Store.AddMount(ctx, alias, path); err != nil {
			return errors.Wrapf(err, "failed to add mount '%s'", alias)
		}
	}

	if backend.HasRCSBackend(ctx) {
		bn := backend.RCSBackendName(backend.GetRCSBackend(ctx))
		out.Debug(ctx, "Initializing RCS (%s) ...", bn)
		if err := s.rcsInit(ctx, alias, GetUsername(ctx), GetEmail(ctx)); err != nil {
			out.Debug(ctx, "Stacktrace: %+v\n", err)
			out.Error(ctx, "Failed to init RCS (%s): %s", bn, err)
		}
	} else {
		out.Debug(ctx, "not initializing RCS backend ...")
	}

	out.Green(ctx, "Password store %s initialized for:", path)
	s.printRecipients(ctx, alias)

	// write config
	if err := s.cfg.Save(); err != nil {
		return ExitError(ctx, ExitConfig, err, "failed to write config: %s", err)
	}

	return nil
}

func (s *Action) printRecipients(ctx context.Context, alias string) {
	crypto := s.Store.Crypto(ctx, alias)
	for _, recipient := range s.Store.ListRecipients(ctx, alias) {
		r := "0x" + recipient
		if kl, err := crypto.FindPublicKeys(ctx, recipient); err == nil && len(kl) > 0 {
			r = crypto.FormatKey(ctx, kl[0])
		}
		out.Yellow(ctx, "  "+r)
	}
}

func (s *Action) getCryptoFor(ctx context.Context, name string) backend.Crypto {
	crypto := s.Store.Crypto(ctx, name)
	if crypto != nil {
		return crypto
	}
	c, err := sub.GetCryptoBackend(ctx, backend.GetCryptoBackend(ctx), config.Directory())
	if err != nil {
		out.Debug(ctx, "getCryptoFor(%s) failed to init crypto backend: %s", name, err)
		return nil
	}
	return c
}

func detectName(c *cli.Context) string {
	for _, e := range []string{
		c.String("name"),
		os.Getenv("GIT_AUTHOR_NAME"),
		os.Getenv("DEBFULLNAME"),
		os.Getenv("USER"),
	} {
		if e != "" {
			return e
		}
	}
	return ""
}
func detectEmail(c *cli.Context) string {
	for _, e := range []string{
		c.String("email"),
		os.Getenv("GIT_AUTHOR_EMAIL"),
		os.Getenv("DEBEMAIL"),
		os.Getenv("EMAIL"),
	} {
		if e != "" {
			return e
		}
	}
	return ""
}

// InitOnboarding will invoke the onboarding / setup wizard
func (s *Action) InitOnboarding(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	remote := c.String("remote")
	team := c.String("alias")
	create := c.Bool("create")
	name := detectName(c)
	if name != "" {
		ctx = WithUsername(ctx, name)
	}
	email := detectEmail(c)
	if email != "" {
		ctx = WithEmail(ctx, email)
	}
	ctx = backend.WithCryptoBackendString(ctx, c.String("crypto"))

	// default to git
	if rcs := c.String("rcs"); rcs != "" {
		ctx = backend.WithRCSBackendString(ctx, c.String("rcs"))
	} else {
		out.Debug(ctx, "Using default RCS backend (GitCLI)")
		ctx = backend.WithRCSBackend(ctx, backend.GitCLI)
	}

	ctx = out.AddPrefix(ctx, "[init] ")
	out.Debug(ctx, "Starting Onboarding Wizard - remote: %s - team: %s - create: %t - name: %s - email: %s", remote, team, create, name, email)

	crypto := s.getCryptoFor(ctx, name)

	out.Debug(ctx, "Crypto Backend initialized as: %s", crypto.Name())

	// check for existing GPG keypairs (private/secret keys). We need at least
	// one useable key pair. If none exists try to create one
	if !s.initHasUseablePrivateKeys(ctx, crypto, team) {
		out.Yellow(ctx, "No useable crypto keys. Generating new key pair")
		ctx := out.AddPrefix(ctx, "[crypto] ")
		out.Print(ctx, "Key generation may take up to a few minutes")
		if err := s.initCreatePrivateKey(ctx, crypto, team, name, email); err != nil {
			return errors.Wrapf(err, "failed to create new private key")
		}
	}

	out.Debug(ctx, "Has useable private keys")

	// if a git remote and a team name are given attempt unattended team setup
	if remote != "" && team != "" {
		if create {
			return s.initCreateTeam(ctx, c, team, remote)
		}
		return s.initJoinTeam(ctx, c, team, remote)
	}

	// no flags given, run interactively
	choices := []string{
		"Local store",
		"Create a Team",
		"Join an existing Team",
	}
	act, sel := cui.GetSelection(ctx, "Select action", "<↑/↓> to change the selection, <→> to select, <ESC> to quit", choices)
	switch act {
	case "default":
		fallthrough
	case "show":
		switch sel {
		case 0:
			return s.initLocal(ctx, c)
		case 1:
			return s.initCreateTeam(ctx, c, "", "")
		case 2:
			return s.initJoinTeam(ctx, c, "", "")
		}
	default:
		return fmt.Errorf("user aborted")
	}
	return nil
}

func (s *Action) initCreatePrivateKey(ctx context.Context, crypto backend.Crypto, mount, name, email string) error {
	out.Green(ctx, "Creating key pair ...")
	out.Yellow(ctx, "WARNING: We are about to generate some GPG keys.")
	out.Print(ctx, `However, the GPG program can sometimes lock up, displaying the following:
"We need to generate a lot of random bytes."
If this happens, please see the following tips:
https://github.com/gopasspw/gopass/blob/master/docs/entropy.md`)
	if name != "" && email != "" {
		ctx := out.AddPrefix(ctx, " ")
		passphrase := xkcdgen.Random()
		if err := crypto.CreatePrivateKeyBatch(ctx, name, email, passphrase); err != nil {
			return errors.Wrapf(err, "failed to create new private key in batch mode")
		}
		out.Green(ctx, "-> OK")
		out.Print(ctx, color.MagentaString("Passphrase: ")+color.HiGreenString(passphrase))
	} else {
		if want, err := termio.AskForBool(ctx, "Continue?", true); err != nil || !want {
			return errors.Wrapf(err, "User aborted")
		}
		ctx := out.WithPrefix(ctx, " ")
		if err := crypto.CreatePrivateKey(ctx); err != nil {
			return errors.Wrapf(err, "failed to create new private key in interactive mode")
		}
		out.Green(ctx, "-> OK")
	}

	kl, err := crypto.ListPrivateKeyIDs(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to list private keys")
	}
	if len(kl) > 1 {
		out.Cyan(ctx, "WARNING: More than one private key detected. Make sure to communicate the right one")
		return nil
	}
	if len(kl) < 1 {
		out.Debug(ctx, "Private Keys: %+v", kl)
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

func (s *Action) initHasUseablePrivateKeys(ctx context.Context, crypto backend.Crypto, mount string) bool {
	kl, err := crypto.ListPrivateKeyIDs(ctx)
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
			out.Debug(ctx, "Initial git pull failed: %s", err)
		}
		if err := s.Store.GitPush(ctx, team, "origin", "master"); err != nil {
			return errors.Wrapf(err, "failed to push to git remote")
		}
	}
	return nil
}

// initLocal will initialize a local store, useful for local-only setups or as
// part of team setups to create the root store
func (s *Action) initLocal(ctx context.Context, c *cli.Context) error {
	ctx = out.AddPrefix(ctx, "[local] ")

	path := ""
	if s.Store != nil {
		path = s.Store.URL()
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
		// autosync
		if want, err := termio.AskForBool(ctx, out.Prefix(ctx)+"Do you want to automatically push any changes to the git remote (if any)?", true); err == nil {
			s.cfg.Root.AutoSync = want
		}
	} else {
		s.cfg.Root.AutoSync = false
	}

	// noconfirm
	if want, err := termio.AskForBool(ctx, out.Prefix(ctx)+"Do you want to always confirm recipients when encrypting?", true); err == nil {
		s.cfg.Root.NoConfirm = !want
	}

	// save config
	if err := s.cfg.Save(); err != nil {
		return errors.Wrapf(err, "failed to save config")
	}

	out.Green(ctx, " -> OK")
	return nil
}

// initCreateTeam will create a local root store and a shared team store
func (s *Action) initCreateTeam(ctx context.Context, c *cli.Context, team, remote string) error {
	var err error

	out.Print(ctx, "Creating a new team ...")
	if err := s.initLocal(ctx, c); err != nil {
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
func (s *Action) initJoinTeam(ctx context.Context, c *cli.Context, team, remote string) error {
	var err error

	out.Print(ctx, "Joining existing team ...")
	if err := s.initLocal(ctx, c); err != nil {
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
