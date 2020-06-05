package action

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/internal/clipboard"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	defaultLength     = 24
	defaultXKCDLength = 4
)

var (
	reNumber = regexp.MustCompile(`^\d+$`)
)

// Generate and save a password
func (s *Action) Generate(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	ctx = WithClip(ctx, c.Bool("clip"))
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
			return ExitError(ExitNoName, err, "please provide a password name")
		}
	}

	// ask for confirmation before overwriting existing entry
	if !force { // don't check if it's force anyway
		if s.Store.Exists(ctx, name) && key == "" && !termio.AskForConfirmation(ctx, fmt.Sprintf("An entry already exists for %s. Overwrite the current password?", name)) {
			return ExitError(ExitAborted, nil, "user aborted. not overwriting your current password")
		}
	}

	// generate password
	password, err := s.generatePassword(ctx, c, length)
	if err != nil {
		return err
	}

	// display or copy to clipboard
	if err := s.generateCopyOrPrint(ctx, c, name, key, password); err != nil {
		return err
	}

	// write generated password to store
	ctx, err = s.generateSetPassword(ctx, name, key, password, kvps)
	if err != nil {
		return err
	}

	// if requested launch editor to add more data to the generated secret
	if edit && termio.AskForConfirmation(ctx, fmt.Sprintf("Do you want to add more data for %s?", name)) {
		c.Context = ctx
		if err := s.Edit(c); err != nil {
			return ExitError(ExitUnknown, err, "failed to edit '%s': %s", name, err)
		}
	}

	return nil
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
	entry := name
	if key != "" {
		entry += ":" + key
	}

	out.Print(ctx, "Password for %s generated", entry)

	if ctxutil.IsAutoClip(ctx) || IsClip(ctx) {
		if err := clipboard.CopyTo(ctx, name, []byte(password)); err != nil {
			return ExitError(ExitIO, err, "failed to copy to clipboard: %s", err)
		}
		if ctxutil.IsAutoClip(ctx) && !c.Bool("print") {
			return nil
		}
	}

	out.Print(
		ctx,
		"The generated password is:\n%s",
		color.YellowString(password),
	)
	return nil
}

// generatePassword will run through the password generation steps
func (s *Action) generatePassword(ctx context.Context, c *cli.Context, length string) (string, error) {
	if c.Bool("xkcd") || c.IsSet("xkcdsep") {
		return s.generatePasswordXKCD(ctx, c, length)
	}

	symbols := false
	if c.IsSet("symbols") {
		symbols = c.Bool("symbols")
	}

	var pwlen int
	if length == "" {
		candidateLength := defaultLength
		question := "How long should the password be?"
		iv, err := termio.AskForInt(ctx, question, candidateLength)
		if err != nil {
			return "", ExitError(ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	} else {
		iv, err := strconv.Atoi(length)
		if err != nil {
			return "", ExitError(ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	}

	if pwlen < 1 {
		return "", ExitError(ExitUsage, nil, "password length must not be zero")
	}

	if c.Bool("strict") {
		return pwgen.GeneratePasswordWithAllClasses(pwlen)
	}
	if c.Bool("memorable") {
		return pwgen.GenerateMemorablePassword(pwlen, symbols), nil
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
			return "", ExitError(ExitUsage, err, "password length must be a number")
		}
		pwlen = iv
	} else {
		iv, err := strconv.Atoi(length)
		if err != nil {
			return "", ExitError(ExitUsage, err, "password length must be a number: %s", err)
		}
		pwlen = iv
	}

	if pwlen < 1 {
		return "", ExitError(ExitUsage, nil, "password length must not be zero")
	}

	return xkcdgen.RandomLengthDelim(pwlen, xkcdSeparator, c.String("xkcdlang"))
}

// generateSetPassword will update or create a secret
func (s *Action) generateSetPassword(ctx context.Context, name, key, password string, kvps map[string]string) (context.Context, error) {
	// set a single key in a yaml doc
	if key != "" {
		gs, err := s.Store.Get(ctx, name)
		if err != nil {
			return ctx, ExitError(ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		sec := gs.MIME()
		setMetadata(sec, kvps)
		sec.Set(key, password)
		if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Generated password for YAML key"), name, sec); err != nil {
			return ctx, ExitError(ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		return ctx, nil
	}

	// replace password in existing secret
	if s.Store.Exists(ctx, name) {
		ctx, err := s.generateReplaceExisting(ctx, name, key, password, kvps)
		if err == nil {
			return ctx, nil
		}
		out.Error(ctx, "Failed to read existing secret. Creating anew. Error: %s", err.Error())
	}

	// generate a completely new secret
	var err error
	var sec gopass.Secret
	sec = secret.New()
	sec.Set("password", password)

	if content, found := s.renderTemplate(ctx, name, []byte(password)); found {
		nSec, err := secparse.Parse(content)
		if err == nil {
			sec = nSec
		} else {
			debug.Log("failed to parse template: %s", err)
		}
	}

	err = s.Store.Set(ctxutil.WithCommitMessage(ctx, "Generated Password"), name, sec)
	if err != nil {
		return ctx, ExitError(ExitEncrypt, err, "failed to create '%s': %s", name, err)
	}
	return ctx, nil
}

func (s *Action) generateReplaceExisting(ctx context.Context, name, key, password string, kvps map[string]string) (context.Context, error) {
	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return ctx, ExitError(ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
	}
	setMetadata(sec, kvps)
	sec.Set("password", password)
	if err := s.Store.Set(ctxutil.WithCommitMessage(ctx, "Generated password for YAML key"), name, sec); err != nil {
		return ctx, ExitError(ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
	}
	return ctx, nil
}

func setMetadata(sec gopass.Secret, kvps map[string]string) {
	for k, v := range kvps {
		sec.Set(k, v)
	}
}

// CompleteGenerate implements the completion heuristic for the generate command
func (s *Action) CompleteGenerate(c *cli.Context) {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.Args().Len() < 1 {
		return
	}
	needle := c.Args().Get(0)

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
