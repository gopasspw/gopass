package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Initialized returns an error if the store is not properly
// prepared.
func (s *Action) Initialized(*cli.Context) error {
	if !s.Store.Initialized() {
		return s.exitError(ExitNotInitialized, nil, "password-store is not initialized. Try '%s init'", s.Name)
	}
	return nil
}

// Init a new password store with a first gpg id
func (s *Action) Init(c *cli.Context) error {
	path := c.String("path")
	alias := c.String("store")
	nogit := c.Bool("nogit")

	if err := s.init(alias, path, nogit, c.Args()...); err != nil {
		return s.exitError(ExitUnknown, err, "failed to initialized store: %s", err)
	}
	return nil
}

func (s *Action) init(alias, path string, nogit bool, keys ...string) error {
	if path == "" {
		path = s.Store.Path()
	}

	if len(keys) < 1 {
		nk, err := s.askForPrivateKey(color.CyanString("Please select a private key for encryption:"))
		if err != nil {
			return errors.Wrapf(err, "failed to read user input")
		}
		keys = []string{nk}
	}

	if err := s.Store.Init(alias, path, keys...); err != nil {
		return errors.Wrapf(err, "failed to init store '%s' at '%s'", alias, path)
	}

	if alias != "" && path != "" {
		if err := s.Store.AddMount(alias, path); err != nil {
			return errors.Wrapf(err, "failed to add mount '%s'", alias)
		}
	}

	if !nogit {
		sk := ""
		if len(keys) == 1 {
			sk = keys[0]
		}
		if err := s.gitInit(alias, sk); err != nil {
			if s.debug {
				fmt.Println(color.RedString("Stacktrace: %+v\n", err))
			}
			fmt.Println(color.RedString("Failed to init git: %s", err))
		}
	}

	fmt.Fprint(color.Output, color.GreenString("Password store %s initialized for:\n", path))
	for _, recipient := range s.Store.ListRecipients(alias) {
		r := "0x" + recipient
		if kl, err := s.gpg.FindPublicKeys(recipient); err == nil && len(kl) > 0 {
			r = kl[0].OneLine()
		}
		fmt.Println(color.YellowString("  " + r))
	}
	fmt.Println("")

	// write config
	if err := s.Store.Config().Save(); err != nil {
		return s.exitError(ExitConfig, err, "failed to write config: %s", err)
	}

	return nil
}
