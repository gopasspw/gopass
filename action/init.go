package action

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/termwiz"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Initialized returns an error if the store is not properly
// prepared.
func (s *Action) Initialized(ctx context.Context, c *cli.Context) error {
	if !s.Store.Initialized() {
		if ctxutil.IsInteractive(ctx) {
			if ok, err := s.askForBool(ctx, "It seems you are new to gopass. Do you want to run the onboarding wizard?", true); err == nil && ok {
				return s.InitOnboarding(ctx, c)
			}
		}
		return s.exitError(ctx, ExitNotInitialized, nil, "password-store is not initialized. Try '%s init'", s.Name)
	}
	return nil
}

// Init a new password store with a first gpg id
func (s *Action) Init(ctx context.Context, c *cli.Context) error {
	path := c.String("path")
	alias := c.String("store")
	nogit := c.Bool("nogit")

	fmt.Println(color.CyanString("Initializing a new password store ...\n"))

	if err := s.init(ctx, alias, path, nogit, c.Args()...); err != nil {
		return s.exitError(ctx, ExitUnknown, err, "failed to initialized store: %s", err)
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
			if ctxutil.IsDebug(ctx) {
				fmt.Println(color.RedString("Stacktrace: %+v\n", err))
			}
			fmt.Println(color.RedString("Failed to init git: %s", err))
		}
	}

	fmt.Fprint(color.Output, color.GreenString("\nPassword store %s initialized for:\n", path))
	for _, recipient := range s.Store.ListRecipients(ctx, alias) {
		r := "0x" + recipient
		if kl, err := s.gpg.FindPublicKeys(ctx, recipient); err == nil && len(kl) > 0 {
			r = kl[0].OneLine()
		}
		fmt.Println(color.YellowString("  " + r))
	}
	fmt.Println("")

	// write config
	if err := s.cfg.Save(); err != nil {
		return s.exitError(ctx, ExitConfig, err, "failed to write config: %s", err)
	}

	return nil
}

// InitOnboarding will invoke the onboarding / setup wizard
func (s *Action) InitOnboarding(ctx context.Context, c *cli.Context) error {
	remote := c.String("remote")
	team := c.String("alias")
	create := c.Bool("create")

	if remote != "" && team != "" {
		if create {
			return s.initOBCreateTeam(ctx, c, team, remote)
		}
		return s.initOBJoinTeam(ctx, c, team, remote)
	}

	// no flags given, ask user
	choices := []string{
		"Local store",
		"Create a Team",
		"Join an existing Team",
	}
	act, sel := termwiz.GetSelection(ctx, "Store for secret", "<↑/↓> to change the selection, <→> to select, <ESC> to quit", choices)
	switch act {
	case "show":
		switch sel {
		case 0:
			return s.initOBLocal(ctx, c)
		case 1:
			return s.initOBCreateTeam(ctx, c, "", "")
		case 2:
			return s.initOBJoinTeam(ctx, c, "", "")
		}
	default:
		return fmt.Errorf("user aborted")
	}
	return nil
}

func (s *Action) initOBLocal(ctx context.Context, c *cli.Context) error {
	fmt.Println("Initializing your local store")
	if err := s.init(ctx, "", "", false); err != nil {
		return err
	}
	fmt.Println("Configuring your local store")
	if want, err := s.askForBool(ctx, "Do you want to automatically push any changes to the git remote (if any)?", true); err == nil {
		s.cfg.Root.AutoSync = want
	}
	if want, err := s.askForBool(ctx, "Do you want to always confirm recipients when encrypting?", false); err == nil {
		s.cfg.Root.NoConfirm = !want
	}
	if err := s.cfg.Save(); err != nil {
		return errors.Wrapf(err, "failed to save config")
	}
	return nil
}

func (s *Action) initOBCreateTeam(ctx context.Context, c *cli.Context, team, remote string) error {
	var err error
	fmt.Println("Ok, creating a new team. We need three things: 1.) a local store for you, 2.) the initial copy of the team store and 3.) a remote to push the store to")
	fmt.Println("1.) Local Store")
	if err := s.initOBLocal(ctx, c); err != nil {
		return errors.Wrapf(err, "failed to create local store")
	}
	team, err = s.askForString(ctx, "Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return err
	}
	fmt.Println("2.) Initializing your shared store for ", team)
	if err := s.init(ctx, team, "", false); err != nil {
		return err
	}
	fmt.Println("3.) Configuring the remote for ", team)
	remote, err = s.askForString(ctx, "Please enter the git remote for your shared store", remote)
	if err != nil {
		return err
	}
	if err := s.Store.Git(ctx, team, false, false, "remote", "add", "origin", remote); err != nil {
		return errors.Wrapf(err, "failed to add git remote")
	}
	if err := s.Store.Git(ctx, team, false, false, "push", "origin", "master"); err != nil {
		return errors.Wrapf(err, "failed to push to git remote")
	}
	return nil
}

func (s *Action) initOBJoinTeam(ctx context.Context, c *cli.Context, team, remote string) error {
	var err error
	fmt.Println("Ok, joining an existing team. We need two things: 1.) a local store for you, 2.) the remote to clone the team store from")
	if err := s.initOBLocal(ctx, c); err != nil {
		return errors.Wrapf(err, "failed to create local store")
	}
	team, err = s.askForString(ctx, "Please enter the name of your team (may contain slashes)", team)
	if err != nil {
		return err
	}
	fmt.Println("2.) Cloning from the remote for ", team)
	remote, err = s.askForString(ctx, "Please enter the git remote for your shared store", remote)
	if err != nil {
		return err
	}
	return s.clone(ctx, remote, team, "")
}
