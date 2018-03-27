package action

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/pkg/clipboard"
	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/pwgen"
	"github.com/justwatchcom/gopass/pkg/pwgen/xkcdgen"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/justwatchcom/gopass/pkg/store/sub"
	"github.com/justwatchcom/gopass/pkg/termio"
	"github.com/urfave/cli"
)

const (
	defaultLength     = 24
	defaultXKCDLength = 4
)

var (
	reNumber = regexp.MustCompile(`^\d+$`)
)

// Generate and save a password
func (s *Action) Generate(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")
	edit := c.Bool("edit")

	name := c.Args().Get(0)
	key, length := keyAndLength(c)

	// ask for name of the secret if it wasn't provided already
	if name == "" {
		var err error
		name, err = termio.AskForString(ctx, "Which name do you want to use?", "")
		if err != nil || name == "" {
			return ExitError(ctx, ExitNoName, err, "please provide a password name")
		}
	}

	// ask for confirmation before overwriting existing entry
	if !force { // don't check if it's force anyway
		if s.Store.Exists(ctx, name) && key == "" && !termio.AskForConfirmation(ctx, fmt.Sprintf("An entry already exists for %s. Overwrite the current password?", name)) {
			return ExitError(ctx, ExitAborted, nil, "user aborted. not overwriting your current password")
		}
	}

	// generate password
	password, err := s.generatePassword(ctx, c, length)
	if err != nil {
		return err
	}

	// write generated password to store
	ctx, err = s.generateSetPassword(ctx, name, key, password)
	if err != nil {
		return err
	}

	// if requested launch editor to add more data to the generated secret
	if (edit || ctxutil.IsAskForMore(ctx)) && termio.AskForConfirmation(ctx, fmt.Sprintf("Do you want to add more data for %s?", name)) {
		if err := s.Edit(ctx, c); err != nil {
			return ExitError(ctx, ExitUnknown, err, "failed to edit '%s': %s", name, err)
		}
	}

	// display or copy to clipboard
	return s.generateCopyOrPrint(ctx, c, name, key, password)
}

func keyAndLength(c *cli.Context) (string, string) {
	key := c.Args().Get(1)
	length := c.Args().Get(2)

	// generate can be called with one positional arg or two
	// one - the desired length for the "master" secret itself
	// two - the key in a YAML doc and the length for a secret generated for this
	// key only
	if length == "" && key != "" && reNumber.MatchString(key) {
		length = key
		key = ""
	}

	return key, length
}

// generateCopyOrPrint will print the password to the screen or copy to the
// clipboard
func (s *Action) generateCopyOrPrint(ctx context.Context, c *cli.Context, name, key, password string) error {
	if c.Bool("print") {
		if key != "" {
			key = " " + key
		}
		out.Print(
			ctx,
			"The generated password for %s%s is:\n%s", name, key,
			color.YellowString(password),
		)
		return nil
	}
	if ctxutil.IsAutoClip(ctx) || c.Bool("clip") {
		if err := clipboard.CopyTo(ctx, name, []byte(password)); err != nil {
			return ExitError(ctx, ExitIO, err, "failed to copy to clipboard: %s", err)
		}
	}
	return nil
}

// generatePassword will run through the password generation steps
func (s *Action) generatePassword(ctx context.Context, c *cli.Context, length string) (string, error) {
	if c.Bool("xkcd") || c.IsSet("xkcdsep") {
		return s.generatePasswordXKCD(ctx, c, length)
	}

	symbols := false
	if c.Bool("symbols") || ctxutil.IsUseSymbols(ctx) {
		symbols = true
	}

	var pwlen int
	if length == "" {
		candidateLength := defaultLength
		question := "How long should the password be?"
		iv, err := termio.AskForInt(ctx, question, candidateLength)
		if err != nil {
			return "", ExitError(ctx, ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	} else {
		iv, err := strconv.Atoi(length)
		if err != nil {
			return "", ExitError(ctx, ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	}

	if pwlen < 1 {
		return "", ExitError(ctx, ExitUsage, nil, "password length must not be zero")
	}

	return pwgen.GeneratePassword(pwlen, symbols), nil
}

// generatePasswordXKCD walks through the steps necessary to create an XKCD-style
// password
func (s *Action) generatePasswordXKCD(ctx context.Context, c *cli.Context, length string) (string, error) {
	xkcdSeparator := " "
	if c.IsSet("xkcdsep") {
		xkcdSeparator = c.String("xkcdsep")
	}

	var pwlen int
	if length == "" {
		candidateLength := defaultXKCDLength
		question := "How many words should be combined to a password?"
		iv, err := termio.AskForInt(ctx, question, candidateLength)
		if err != nil {
			return "", ExitError(ctx, ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	} else {
		iv, err := strconv.Atoi(length)
		if err != nil {
			return "", ExitError(ctx, ExitUsage, err, "password length must be a number: %s", err)
		}
		pwlen = iv
	}

	if pwlen < 1 {
		return "", ExitError(ctx, ExitUsage, nil, "password length must not be zero")
	}

	return xkcdgen.RandomLengthDelim(pwlen, xkcdSeparator, c.String("xkcdlang"))
}

// generateSetPassword will update or create a secret
func (s *Action) generateSetPassword(ctx context.Context, name, key, password string) (context.Context, error) {
	// set a single key in a yaml doc
	if key != "" {
		sec, ctx, err := s.Store.GetContext(ctx, name)
		if err != nil {
			return ctx, ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		if err := sec.SetValue(key, password); err != nil {
			return ctx, ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		if err := s.Store.Set(sub.WithReason(ctx, "Generated password for YAML key"), name, sec); err != nil {
			return ctx, ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		return ctx, nil
	}

	// replace password in existing secret
	if s.Store.Exists(ctx, name) {
		sec, ctx, err := s.Store.GetContext(ctx, name)
		if err != nil {
			return ctx, ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		sec.SetPassword(password)
		if err := s.Store.Set(sub.WithReason(ctx, "Generated password for YAML key"), name, sec); err != nil {
			return ctx, ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		return ctx, nil
	}

	// generate a completely new secret
	var err error
	ctx, err = s.Store.SetContext(sub.WithReason(ctx, "Generated Password"), name, secret.New(password, ""))
	if err != nil {
		return ctx, ExitError(ctx, ExitEncrypt, err, "failed to create '%s': %s", name, err)
	}
	return ctx, nil
}
