package action

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/pwgen"
	"github.com/urfave/cli"
)

const (
	defaultLength = 24
)

// Generate & save a password
func (s *Action) Generate(ctx context.Context, c *cli.Context) error {
	force := c.Bool("force")
	edit := c.Bool("edit")
	symbols := c.Bool("symbols")
	if c.IsSet("no-symbols") {
		fmt.Println(color.RedString("Warning: -n/--no-symbols is deprecated. This is now the default. Use -s to enable symbols"))
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
		name, err = s.askForString("Which name do you want to use?", "")
		if err != nil || name == "" {
			return s.exitError(ctx, ExitNoName, err, "please provide a password name")
		}
	}

	if !force { // don't check if it's force anyway
		if s.Store.Exists(name) && key == "" && !s.AskForConfirmation(ctx, fmt.Sprintf("An entry already exists for %s. Overwrite the current password?", name)) {
			return s.exitError(ctx, ExitAborted, nil, "user aborted. not overwriting your current password")
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
		return s.exitError(ctx, ExitUsage, err, "password length must be a number")
	}
	if pwlen < 1 {
		return s.exitError(ctx, ExitUsage, nil, "password length must not be zero")
	}

	password := pwgen.GeneratePassword(pwlen, symbols)

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
	} else if s.Store.Exists(name) {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		sec.SetPassword(string(password))
		if err := s.Store.Set(sub.WithReason(ctx, "Generated password for YAML key"), name, sec); err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
	} else {
		if err := s.Store.Set(sub.WithReason(ctx, "Generated Password"), name, secret.New(string(password), "")); err != nil {
			return s.exitError(ctx, ExitEncrypt, err, "failed to create '%s': %s", name, err)
		}
	}

	if c.Bool("clip") {
		return s.copyToClipboard(ctx, name, password)
	}

	if key != "" {
		key = " " + key
	}
	fmt.Printf(
		"The generated password for %s%s is:\n%s\n", name, key,
		color.YellowString(string(password)),
	)

	if (edit || ctxutil.IsAskForMore(ctx)) && s.AskForConfirmation(ctx, fmt.Sprintf("Do you want to add more data for %s?", name)) {
		if err := s.Edit(ctx, c); err != nil {
			return s.exitError(ctx, ExitUnknown, err, "failed to edit '%s': %s", name, err)
		}
	}

	return nil
}
