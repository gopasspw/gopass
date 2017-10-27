package action

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/pwgen"
	"github.com/justwatchcom/gopass/utils/pwgen/xkcdgen"
	"github.com/urfave/cli"
)

const (
	defaultLength     = 24
	defaultXKCDLength = 4
)

// Generate & save a password
func (s *Action) Generate(ctx context.Context, c *cli.Context) error {
	var xkcdSeparator string
	force := c.Bool("force")
	edit := c.Bool("edit")
	symbols := false
	if c.Bool("symbols") || ctxutil.IsUseSymbols(ctx) {
		symbols = true
	}
	xkcd := c.Bool("xkcd")
	if c.IsSet("xkcdsep") {
		xkcdSeparator = c.String("xkcdsep")
		xkcd = true
	}

	if c.IsSet("no-symbols") {
		out.Red(ctx, "Warning: -n/--no-symbols is deprecated. This is now the default. Use -s to enable symbols. You can also set 'usesymbols' to true via gopass config.")
	}

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
		name, err = s.askForString(ctx, "Which name do you want to use?", "")
		if err != nil || name == "" {
			return s.exitError(ctx, ExitNoName, err, "please provide a password name")
		}
	}

	if !force { // don't check if it's force anyway
		if s.Store.Exists(ctx, name) && key == "" && !s.AskForConfirmation(ctx, fmt.Sprintf("An entry already exists for %s. Overwrite the current password?", name)) {
			return s.exitError(ctx, ExitAborted, nil, "user aborted. not overwriting your current password")
		}
	}

	var pwlen int
	if length == "" {
		candidateLength := defaultLength
		question := "How long should the password be?"
		if xkcd {
			candidateLength = defaultXKCDLength
			question = "How many words should be combined to a password?"
		}
		iv, err := s.askForInt(ctx, question, candidateLength)
		if err != nil {
			return s.exitError(ctx, ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	} else {
		iv, err := strconv.Atoi(length)
		if err != nil {
			return s.exitError(ctx, ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	}

	if pwlen < 1 {
		return s.exitError(ctx, ExitUsage, nil, "password length must not be zero")
	}

	var password string
	if xkcd {
		var err error
		password, err = xkcdgen.RandomLengthDelim(pwlen, xkcdSeparator, c.String("xkcdlang"))
		if err != nil {
			return err
		}
	} else {
		password = pwgen.GeneratePassword(pwlen, symbols)
	}

	// set a single key in a yaml doc
	if key != "" {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		if err := sec.SetValue(key, string(password)); err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		if err := s.Store.Set(sub.WithReason(ctx, "Generated password for YAML key"), name, sec); err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
	} else if s.Store.Exists(ctx, name) {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		sec.SetPassword(password)
		if err := s.Store.Set(sub.WithReason(ctx, "Generated password for YAML key"), name, sec); err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
	} else {
		if err := s.Store.Set(sub.WithReason(ctx, "Generated Password"), name, secret.New(string(password), "")); err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to create '%s': %s", name, err)
		}
	}

	if (edit || ctxutil.IsAskForMore(ctx)) && s.AskForConfirmation(ctx, fmt.Sprintf("Do you want to add more data for %s?", name)) {
		if err := s.Edit(ctx, c); err != nil {
			return s.exitError(ctx, ExitUnknown, err, "failed to edit '%s': %s", name, err)
		}
	}

	if c.Bool("print") {
		if key != "" {
			key = " " + key
		}
		fmt.Printf(
			"The generated password for %s%s is:\n%s\n", name, key,
			color.YellowString(string(password)),
		)
		return nil
	}

	return s.copyToClipboard(ctx, name, []byte(password))
}
