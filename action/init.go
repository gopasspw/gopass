package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/gpg"
	"github.com/mattn/go-colorable"
	"github.com/urfave/cli"
)

var (
	out = colorable.NewColorableStdout()
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
	store := c.String("store")
	nogit := c.Bool("nogit")

	if !hasConfig() {
		// when creating a new config we set some sensible defaults
		s.Store.AutoPush = true
		s.Store.AutoPull = true
		s.Store.AutoImport = false
		s.Store.NoConfirm = false
		s.Store.PersistKeys = true
		s.Store.LoadKeys = false
		s.Store.ClipTimeout = 45
	}

	keys := c.Args()
	if len(keys) < 1 {
		nk, err := askForPrivateKey("Please select a private Key for encryption:")
		if err != nil {
			return err
		}
		keys = []string{nk}
	}

	if err := s.Store.Init(store, keys...); err != nil {
		return err
	}

	color.Green("Password store initialized for: ")
	for _, recipient := range s.Store.ListRecipients(store) {
		r := "0x" + recipient
		if kl, err := gpg.ListPublicKeys(recipient); err == nil && len(kl) > 0 {
			r = kl[0].OneLine()
		}
		color.Yellow(r)
	}
	fmt.Println("")

	// write config
	if err := writeConfig(s.Store); err != nil {
		color.Red(fmt.Sprintf("Failed to write config: %s", err))
	}

	if nogit {
		return nil
	}

	return s.GitInit(c)
}
