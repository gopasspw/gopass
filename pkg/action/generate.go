package action

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"
	"github.com/gopasspw/gopass/pkg/store"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/pkg/store/sub"
	"github.com/gopasspw/gopass/pkg/termio"

	"github.com/fatih/color"
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

	args, kvps := parseArgs(c)
	name := args.Get(0)
	key, length := keyAndLength(args)

	ctx = ctxutil.WithForce(ctx, force)

	// ask for name of the secret if it wasn't provided already
	if name == "" {
		var err error
		name, err = termio.AskForString(ctx, "Which name do you want to use?", "")
		if err != nil || name == "" {
			return ExitError(ctx, ExitNoName, err, "please provide a password name")
		}
	}

	ctx = s.Store.WithConfig(ctx, name)

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
	ctx, err = s.generateSetPassword(ctx, name, key, password, kvps)
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

func keyAndLength(args argList) (string, string) {
	key := args.Get(1)
	length := args.Get(2)

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
	if ctxutil.IsAutoPrint(ctx) || c.Bool("print") {
		if key != "" {
			key = " " + key
		}
		out.Print(
			ctx,
			"The generated password for %s%s is:\n%s", name, key,
			color.YellowString(password),
		)
	}

	if ctxutil.IsAutoClip(ctx) || c.Bool("clip") {
		if err := clipboard.CopyTo(ctx, name, []byte(password)); err != nil {
			return ExitError(ctx, ExitIO, err, "failed to copy to clipboard: %s", err)
		}
	}

	if c.Bool("print") || c.Bool("clip") {
		return nil
	}

	entry := name
	if key != "" {
		entry += ":" + key
	}
	out.Print(ctx, "Password for %s generated", entry)
	return nil
}

// generatePassword will run through the password generation steps
func (s *Action) generatePassword(ctx context.Context, c *cli.Context, length string) (string, error) {
	if c.Bool("xkcd") || c.IsSet("xkcdsep") {
		return s.generatePasswordXKCD(ctx, c, length)
	}

	symbols := ctxutil.IsUseSymbols(ctx)
	if c.IsSet("symbols") {
		symbols = c.Bool("symbols")
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

	corp, err := termio.AskForBool(ctx, "Do you have strict rules to include different character classes?", false)
	if err != nil {
		return "", err
	}
	if corp {
		return pwgen.GeneratePasswordWithAllClasses(pwlen)
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
func (s *Action) generateSetPassword(ctx context.Context, name, key, password string, kvps map[string]string) (context.Context, error) {
	// set a single key in a yaml doc
	if key != "" {
		sec, ctx, err := s.Store.GetContext(ctx, name)
		if err != nil {
			return ctx, ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		setMetadata(sec, kvps)
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
		setMetadata(sec, kvps)
		sec.SetPassword(password)
		if err := s.Store.Set(sub.WithReason(ctx, "Generated password for YAML key"), name, sec); err != nil {
			return ctx, ExitError(ctx, ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		return ctx, nil
	}

	// generate a completely new secret
	var err error
	sec := secret.New(password, "")

	if content, found := s.renderTemplate(ctx, name, []byte(password)); found {
		nSec, err := secret.Parse(content)
		if err == nil {
			sec = nSec
		}
	}

	ctx, err = s.Store.SetContext(sub.WithReason(ctx, "Generated Password"), name, sec)
	if err != nil {
		return ctx, ExitError(ctx, ExitEncrypt, err, "failed to create '%s': %s", name, err)
	}
	return ctx, nil
}

func setMetadata(sec store.Secret, kvps map[string]string) {
	for k, v := range kvps {
		_ = sec.SetValue(k, v)
	}
}

// CompleteGenerate implements the completion heuristic for the generate command
func (s *Action) CompleteGenerate(ctx context.Context, c *cli.Context) {
	args := c.Args()
	if len(args) < 1 {
		return
	}
	needle := args[0]

	_, err := s.Store.Initialized(ctx) // important to make sure the structs are not nil
	if err != nil {
		out.Error(ctx, "Store not initialized: %s", err)
		return
	}
	list, err := s.Store.List(ctx, 0)
	if err != nil {
		return
	}

	if strings.Contains(needle, "/") {
		list = filterPrefix(uniq(extractEmails(list)), path.Base(needle))
	} else {
		list = filterPrefix(uniq(extractDomains(list)), needle)
	}

	for _, v := range list {
		fmt.Fprintln(stdout, bashEscape(v))
	}
}

func extractEmails(list []string) []string {
	results := make([]string, 0, len(list))
	for _, e := range list {
		e = path.Base(e)
		if strings.Contains(e, "@") || strings.Contains(e, "_") {
			results = append(results, e)
		}
	}
	return results
}

var reDomain = regexp.MustCompile(`^(?i)([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)

func extractDomains(list []string) []string {
	results := make([]string, 0, len(list))
	for _, e := range list {
		e = path.Base(e)
		if reDomain.MatchString(e) {
			results = append(results, e)
		}
	}
	return results
}

func uniq(in []string) []string {
	set := make(map[string]struct{}, len(in))
	for _, e := range in {
		set[e] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func filterPrefix(in []string, prefix string) []string {
	out := make([]string, 0, len(in))
	for _, e := range in {
		if strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	return out
}
