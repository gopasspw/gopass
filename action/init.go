package action

import (
	"context"
	"fmt"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/pwgen/xkcdgen"
	"github.com/justwatchcom/gopass/utils/termwiz"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Initialized returns an error if the store is not properly
// prepared.
func (s *Action) Initialized(ctx context.Context, c *cli.Context) error {
	if s.gpg.Binary() == "" {
		return exitError(ctx, ExitGPG, nil, "gpg not found but required")
	}
	if s.gpg.Version(ctx).LT(semver.Version{Major: 2, Minor: 0, Patch: 0}) {
		out.Red(ctx, "Warning: Using GPG 1.x. Using GPG 2.0 or later is highly recommended")
	}
	if !s.Store.Initialized() {
		if ctxutil.IsInteractive(ctx) {
			if ok, err := s.askForBool(ctx, "It seems you are new to gopass. Do you want to run the onboarding wizard?", true); err == nil && ok {
				if err := s.InitOnboarding(ctx, c); err != nil {
					return exitError(ctx, ExitUnknown, err, "failed to run onboarding wizard: %s", err)
				}
				return nil
			}
		}
		return exitError(ctx, ExitNotInitialized, nil, "password-store is not initialized. Try '%s init'", s.Name)
	}
	return nil
}

// Init a new password store with a first gpg id
func (s *Action) Init(ctx context.Context, c *cli.Context) error {
	path := c.String("path")
	alias := c.String("store")
	nogit := c.Bool("nogit")

	ctx = out.WithPrefix(ctx, "[init] ")
	out.Cyan(ctx, "Initializing a new password store ...")

	if err := s.init(ctx, alias, path, nogit, c.Args()...); err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to initialized store: %s", err)
	}
	return nil
}

func (s *Action) init(ctx context.Context, alias, path string, nogit bool, keys ...string) error {
	if path == "" {
		if alias != "" {
			path = config.PwStoreDir(alias)
		} else {
			path = s.Store.Path()
		}
	}

	if len(keys) < 1 {
		nk, err := s.askForPrivateKey(ctx, color.CyanString("Please select a private key for encrypting secrets:"))
		if err != nil {
			return errors.Wrapf(err, "failed to read user input")
		}
		keys = []string{nk}
	}

	if err := s.Store.Init(ctx, alias, path, keys...); err != nil {
		return errors.Wrapf(err, "failed to init store '%s' at '%s'", alias, path)
	}

	if alias != "" && path != "" {
		if err := s.Store.AddMount(ctx, alias, path); err != nil {
			return errors.Wrapf(err, "failed to add mount '%s'", alias)
		}
	}

	if !nogit {
		sk := ""
		if len(keys) == 1 {
			sk = keys[0]
		}
		if err := s.gitInit(ctx, alias, sk); err != nil {
			out.Debug(ctx, "Stacktrace: %+v\n", err)
			out.Red(ctx, "Failed to init git: %s", err)
		}
	}

	out.Green(ctx, "Password store %s initialized for:", path)
	for _, recipient := range s.Store.ListRecipients(ctx, alias) {
		r := "0x" + recipient
		if kl, err := s.gpg.FindPublicKeys(ctx, recipient); err == nil && len(kl) > 0 {
			r = kl[0].OneLine()
		}
		out.Yellow(ctx, "  "+r)
	}

	// write config
	if err := s.cfg.Save(); err != nil {
		return exitError(ctx, ExitConfig, err, "failed to write config: %s", err)
	}

	return nil
}

// InitOnboarding will invoke the onboarding / setup wizard
func (s *Action) InitOnboarding(ctx context.Context, c *cli.Context) error {
	remote := c.String("remote")
	team := c.String("alias")
	create := c.Bool("create")
	name := c.String("name")
	email := c.String("email")

	ctx = out.AddPrefix(ctx, "[init] ")

	// check for existing GPG keypairs (private/secret keys). We need at least
	// one useable key pair. If none exists try to create one
	if !s.initHasUseablePrivateKeys(ctx) {
		out.Yellow(ctx, "No useable GPG keys. Generating new key pair")
		ctx := out.AddPrefix(ctx, "[gpg] ")
		out.Print(ctx, "Key generation may take up to a few minutes")
		if err := s.initCreatePrivateKey(ctx, name, email); err != nil {
			return errors.Wrapf(err, "failed to create new private key")
		}
	}

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
	act, sel := termwiz.GetSelection(ctx, "Select action", "<↑/↓> to change the selection, <→> to select, <ESC> to quit", choices)
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

func (s *Action) initCreatePrivateKey(ctx context.Context, name, email string) error {
	out.Green(ctx, "Creating key pair ...")
	out.Yellow(ctx, "WARNING: We are about to generate some GPG keys.")
	out.Print(ctx, `However, the GPG program can sometimes lock up, displaying the following:
"We need to generate a lot of random bytes."
If this happens, please see the following tips:
https://github.com/justwatchcom/gopass/blob/master/docs/entropy.md`)
	if name != "" && email != "" {
		ctx := out.AddPrefix(ctx, " ")
		passphrase := xkcdgen.Random()
		if err := s.gpg.CreatePrivateKeyBatch(ctx, name, email, passphrase); err != nil {
			return errors.Wrapf(err, "failed to create new private key in batch mode")
		}
		out.Green(ctx, "-> OK")
		out.Print(ctx, color.MagentaString("Passphrase: ")+color.HiGreenString(passphrase))
	} else {
		if want, err := s.askForBool(ctx, "Continue?", true); err != nil || !want {
			return errors.Wrapf(err, "User aborted")
		}
		ctx := out.WithPrefix(ctx, " ")
		if err := s.gpg.CreatePrivateKey(ctx); err != nil {
			return errors.Wrapf(err, "failed to create new private key in interactive mode")
		}
		out.Green(ctx, "-> OK")
	}

	kl, err := s.gpg.ListPrivateKeys(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to list private keys")
	}
	klu := kl.UseableKeys()
	if len(klu) > 1 {
		out.Cyan(ctx, "WARNING: More than one private key detected. Make sure to communicate the right one")
		return nil
	}
	if len(klu) < 1 {
		out.Debug(ctx, "Private Keys: %+v", kl)
		return errors.New("failed to create a useable key pair")
	}
	key := klu[0]
	fn := key.ID() + ".pub.key"
	if err := s.gpg.ExportPublicKey(ctx, key.Fingerprint, fn); err != nil {
		return errors.Wrapf(err, "failed to export public key")
	}
	out.Cyan(ctx, "Public key exported to '%s'", fn)
	out.Green(ctx, "Done")
	return nil
}

func (s *Action) initHasUseablePrivateKeys(ctx context.Context) bool {
	kl, err := s.gpg.ListPrivateKeys(ctx)
	if err != nil {
		return false
	}
	return len(kl.UseableKeys()) > 0
}

func (s *Action) initSetupGitRemote(ctx context.Context, team, remote string) error {
	var err error
	remote, err = s.askForString(ctx, "Please enter the git remote for your shared store", remote)
	if err != nil {
		return errors.Wrapf(err, "failed to read user input")
	}
	{
		ctx := out.WithHidden(ctx, true)
		if err := s.Store.Git(ctx, team, false, false, "remote", "add", "origin", remote); err != nil {
			return errors.Wrapf(err, "failed to add git remote")
		}
		// initial pull, in case the remote is non-empty
		if err := s.Store.Git(ctx, team, false, false, "pull", "origin", "master"); err != nil {
			out.Debug(ctx, "Initial git pull failed: %s", err)
		}
		if err := s.Store.Git(ctx, team, false, false, "push", "origin", "master"); err != nil {
			return errors.Wrapf(err, "failed to push to git remote")
		}
	}
	return nil
}

// initLocal will initialize a local store, useful for local-only setups or as
// part of team setups to create the root store
func (s *Action) initLocal(ctx context.Context, c *cli.Context) error {
	ctx = out.AddPrefix(ctx, "[local] ")

	out.Print(ctx, "Initializing your local store ...")
	out.Yellow(ctx, "Setting up git to sign commits. You will be asked for your selected GPG keys passphrase to sign the initial commit")
	if err := s.init(out.WithHidden(ctx, true), "", "", false); err != nil {
		return errors.Wrapf(err, "failed to init local store")
	}
	out.Green(ctx, " -> OK")

	out.Print(ctx, "Configuring your local store ...")

	if want, err := s.askForBool(ctx, "Do you want to add a git remote?", false); err == nil && want {
		out.Print(ctx, "Configuring the git remote ...")
		if err := s.initSetupGitRemote(ctx, "", ""); err != nil {
			return errors.Wrapf(err, "failed to setup git remote")
		}
		// autosync
		if want, err := s.askForBool(ctx, "Do you want to automatically push any changes to the git remote (if any)?", true); err == nil {
			s.cfg.Root.AutoSync = want
		}
	} else {
		s.cfg.Root.AutoSync = false
	}

	// noconfirm
	if want, err := s.askForBool(ctx, "Do you want to always confirm recipients when encrypting?", false); err == nil {
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
	team, err = s.askForString(ctx, "Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return errors.Wrapf(err, "failed to read user input")
	}
	ctx = out.AddPrefix(ctx, "["+team+"] ")

	out.Print(ctx, "Initializing your shared store ...")
	if err := s.init(out.WithHidden(ctx, true), team, "", false); err != nil {
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
	team, err = s.askForString(ctx, "Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return err
	}
	ctx = out.AddPrefix(ctx, "["+team+"]")

	out.Print(ctx, "Configuring git remote ...")
	remote, err = s.askForString(ctx, "Please enter the git remote for your shared store", remote)
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
