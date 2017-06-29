package action

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/pwgen"
	"github.com/urfave/cli"
)

const (
	defaultLength = 24
)

// Generate & save a password
func (s *Action) Generate(c *cli.Context) error {
	force := c.Bool("force")
	noSymbols := c.Bool("no-symbols")

	name := c.Args().Get(0)
	key := c.Args().Get(1)
	length := c.Args().Get(2)

	// generate can be called with one positional arg or two
	// one - the desired length for the "master" secret itself
	// two - the key in a YAML doc and the length for a secret generated for this
	// key only
	if length == "" && key != "" {
		length = key
		key = ""
	}

	if name == "" {
		var err error
		name, err = s.askForString("Which name do you want to use?", "")
		if err != nil || name == "" {
			return fmt.Errorf(color.RedString("provide a password name"))
		}
	}

	if !force { // don't check if it's force anyway
		if s.Store.Exists(name) && key == "" && !s.askForConfirmation(fmt.Sprintf("An entry already exists for %s. Overwrite the current password?", name)) {
			return fmt.Errorf("not overwriting your current password")
		}
	}

	if length == "" {
		length = strconv.Itoa(defaultLength)
		if l, err := s.askForInt("How long should the password be?", defaultLength); err == nil {
			length = strconv.Itoa(l)
		}
	}

	pwlen, err := strconv.Atoi(length)
	if err != nil {
		return fmt.Errorf("password length must be a number")
	}
	if pwlen < 1 {
		return fmt.Errorf("password length must be bigger than 0")
	}

	password := pwgen.GeneratePassword(pwlen, !noSymbols)

	// set a single key in a yaml doc
	if key != "" {
		if err := s.Store.SetKey(name, key, string(password)); err != nil {
			return err
		}
	} else if s.Store.Exists(name) {
		if err := s.Store.SetPassword(name, password); err != nil {
			return err
		}
	} else {
		if err := s.Store.SetConfirm(name, password, "Generated Password", s.confirmRecipients); err != nil {
			return err
		}
	}

	if c.Bool("clip") {
		return s.copyToClipboard(name, password)
	}

	if key != "" {
		key = " " + key
	}
	fmt.Printf(
		"The generated password for %s%s is:\n%s\n", name, key,
		color.YellowString(string(password)),
	)

	if s.Store.AskForMore() && s.askForConfirmation(fmt.Sprintf("Do you want to add more data for %s?", name)) {
		return s.Edit(c)
	}

	return nil
}
