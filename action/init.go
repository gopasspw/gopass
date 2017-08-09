package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Initialized returns an error if the store is not properly
// prepared.
func (s *Action) Initialized(*cli.Context) error {
	if !s.Store.Initialized() {
		return fmt.Errorf("password-store is not initialized. Try '%s init'", s.Name)
	}
	return nil
}

// Init a new password store with a first gpg id
func (s *Action) Init(c *cli.Context) error {
	path := c.String("path")
	alias := c.String("store")
	nogit := c.Bool("nogit")

	return s.init(alias, path, nogit, c.Args()...)
}

func (s *Action) init(alias, path string, nogit bool, keys ...string) error {
	if path == "" {
		path = s.Store.Path()
	}

	if len(keys) < 1 {
		nk, err := s.askForPrivateKey(color.CyanString("Please select a private key for encryption:"))
		if err != nil {
			return err
		}
		keys = []string{nk}
	}

	if err := s.Store.Init(alias, path, keys...); err != nil {
		return err
	}

	if alias != "" && path != "" {
		if err := s.Store.AddMount(alias, path); err != nil {
			return err
		}
	}

	if !nogit {
		sk := ""
		if len(keys) == 1 {
			sk = keys[0]
		}
		if err := s.gitInit(alias, sk); err != nil {
			color.Yellow("Failed to init git: %s", err)
		}
	}

	fmt.Fprint(color.Output, color.GreenString("Password store %s initialized for:\n", path))
	for _, recipient := range s.Store.ListRecipients(alias) {
		r := "0x" + recipient
		if kl, err := s.gpg.FindPublicKeys(recipient); err == nil && len(kl) > 0 {
			r = kl[0].OneLine()
		}
		color.Yellow("  " + r)
	}
	fmt.Println("")

	// write config
	if err := s.Store.Config().Save(); err != nil {
		color.Red(fmt.Sprintf("Failed to write config: %s", err))
	}

	return nil
}
