package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/gpg"
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
	path := c.String("store")
	alias := c.String("alias")
	nogit := c.Bool("nogit")

	return s.init(alias, path, nogit, c.Args()...)
}

func (s *Action) init(alias, path string, nogit bool, keys ...string) error {
	if path != "" && alias == "" {
		return fmt.Errorf("need mount alias when using path")
	}
	if !hasConfig() {
		// when creating a new config we set some sensible defaults
		s.Store.AutoPush = true
		s.Store.AutoPull = true
		s.Store.AutoImport = false
		s.Store.NoConfirm = false
		s.Store.PersistKeys = true
		s.Store.LoadKeys = false
		s.Store.ClipTimeout = 45
		s.Store.ShowSafeContent = false
	}
	if path == "" {
		path = s.Store.Path
	}

	if len(keys) < 1 {
		nk, err := askForPrivateKey(color.CyanString("Please select a private key for encryption:"))
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
		if err := s.gitInit(alias, ""); err != nil {
			color.Yellow("Failed to init git: %s", err)
		}
	}

	fmt.Fprint(color.Output, color.GreenString("Password store %s initialized for:\n", path))
	for _, recipient := range s.Store.ListRecipients(alias) {
		r := "0x" + recipient
		if kl, err := gpg.ListPublicKeys(recipient); err == nil && len(kl) > 0 {
			r = kl[0].OneLine()
		}
		color.Yellow("  " + r)
	}
	fmt.Println("")

	// write config
	if err := writeConfig(s.Store); err != nil {
		color.Red(fmt.Sprintf("Failed to write config: %s", err))
	}

	return nil
}
