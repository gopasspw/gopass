package action

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Initialized returns an error if the store is not properly
// prepared.
func (s *Action) Initialized(ctx context.Context, c *cli.Context) error {
	if !s.Store.Initialized() {
		return s.exitError(ctx, ExitNotInitialized, nil, "password-store is not initialized. Try '%s init'", s.Name)
	}
	return nil
}

// Init a new password store with a first gpg id
func (s *Action) Init(ctx context.Context, c *cli.Context) error {
	path := c.String("path")
	alias := c.String("store")
	nogit := c.Bool("nogit")

	if err := s.init(ctx, alias, path, nogit, c.Args()...); err != nil {
		return s.exitError(ctx, ExitUnknown, err, "failed to initialized store: %s", err)
	}
	return nil
}

func (s *Action) init(ctx context.Context, alias, path string, nogit bool, keys ...string) error {
	if path == "" {
		path = s.Store.Path()
	}

	if len(keys) < 1 {
		nk, err := s.askForPrivateKey(ctx, color.CyanString("Please select a private key for encryption:"))
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

	fmt.Fprint(color.Output, color.GreenString("Password store %s initialized for:\n", path))
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
