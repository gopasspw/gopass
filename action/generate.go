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
	length := c.Args().Get(1)

	if name == "" {
		var err error
		name, err = askForString("Which name do you want to use?", "")
		if err != nil || name == "" {
			return fmt.Errorf(color.RedString("provide a password name"))
		}
	}

	if !force { // don't check if it's force anyway
		if s.Store.Exists(name) && !askForConfirmation(fmt.Sprintf("An entry already exists for %s. Overwrite it?", name)) {
			return fmt.Errorf("not overwriting your current password")
		}
	}

	if length == "" {
		length = strconv.Itoa(defaultLength)
		if l, err := askForInt("How long should the password be?", defaultLength); err == nil {
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

	if err := s.Store.SetConfirm(name, password, "Generated Password", s.confirmRecipients); err != nil {
		return err
	}

	if c.Bool("clip") {
		return s.copyToClipboard(name, password)
	}

	fmt.Printf(
		"The generated password for %s is:\n%s\n", name,
		color.YellowString(string(password)),
	)

	if s.Store.AskForMore() && askForConfirmation(fmt.Sprintf("Do you want to add more data for %s?", name)) {
		return s.Edit(c)
	}

	return nil
}
