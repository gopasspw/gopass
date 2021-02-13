package create

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	"github.com/urfave/cli/v2"
)

const (
	defaultLength = 24
)

type storer interface {
	Set(context.Context, string, gopass.Byter) error
	Exists(context.Context, string) bool
	MountPoints() []string
}

type creator struct {
	store storer
}

func fmtfn(d int, n string, t string) string {
	strlen := 40 - d
	return fmt.Sprintf("%"+strconv.Itoa(d)+"s%s %-"+strconv.Itoa(strlen)+"s", "", color.GreenString("["+n+"]"), color.CyanString(t))
}

// Create displays the password creation wizard
func Create(c *cli.Context, store storer) error {
	ctx := ctxutil.WithGlobalFlags(c)
	s := creator{store: store}
	acts := make(cui.Actions, 0, 5)
	acts = append(acts, cui.Action{Name: "Website Login", Fn: s.createWebsite})
	acts = append(acts, cui.Action{Name: "PIN Code (numerical)", Fn: s.createPIN})
	acts = append(acts, cui.Action{Name: "Generic", Fn: s.createGeneric})
	act, sel := cui.GetSelection(ctx, "Please select the type of secret you would like to create", acts.Selection())
	switch act {
	case "default":
		fallthrough
	case "show":
		return acts.Run(ctx, c, sel)
	default:
		return action.ExitError(action.ExitAborted, nil, "user aborted")
	}
}

// extractHostname tries to extract the hostname from a URL in a filepath-safe
// way for use in the name of a secret
func extractHostname(in string) string {
	if in == "" {
		return ""
	}
	// help url.Parse by adding a scheme if one is missing. This should still
	// allow for any scheme, but by default we assume http (only for parsing)
	urlStr := in
	if !strings.Contains(urlStr, "://") {
		urlStr = "http://" + urlStr
	}
	u, err := url.Parse(urlStr)
	if err == nil {
		if ch := fsutil.CleanFilename(u.Hostname()); ch != "" {
			return ch
		}
	}
	return fsutil.CleanFilename(in)
}

// createWebsite walks through the website credential creation wizard
func (s *creator) createWebsite(ctx context.Context, c *cli.Context) error {
	var (
		urlStr   = c.Args().Get(0)
		username = c.Args().Get(1)
		password string
		comment  string
		store    = c.String("store")
		err      error
		genPw    bool
	)
	out.Printf(ctx, "=> Creating Website login")
	urlStr, err = termio.AskForString(ctx, fmtfn(2, "1", "URL"), urlStr)
	if err != nil {
		return err
	}
	// the hostname is used as part of the name
	hostname := extractHostname(urlStr)
	if hostname == "" {
		return action.ExitError(action.ExitUnknown, err, "Can not parse URL %q. Please use 'gopass edit' to manually create the secret", urlStr)
	}

	username, err = termio.AskForString(ctx, fmtfn(2, "2", "Login"), username)
	if err != nil {
		return err
	}

	genPw, err = termio.AskForBool(ctx, fmtfn(2, "3", "Generate Password?"), true)
	if err != nil {
		return err
	}

	if genPw {
		password, err = s.createGeneratePassword(ctx, hostname)
		if err != nil {
			return err
		}
	} else {
		password, err = termio.AskForPassword(ctx, username)
		if err != nil {
			return err
		}
	}
	comment, _ = termio.AskForString(ctx, fmtfn(2, "4", "Comments"), "")

	// select store
	if store == "" {
		store = cui.AskForStore(ctx, s.store)
	}

	// generate name, ask for override if already taken
	if store != "" {
		store += "/"
	}

	name := fmt.Sprintf("%swebsites/%s/%s", store, fsutil.CleanFilename(hostname), fsutil.CleanFilename(username))
	if s.store.Exists(ctx, name) {
		name, err = termio.AskForString(ctx, fmtfn(2, "5", "Secret already exists, please choose another path"), name)
		if err != nil {
			return err
		}
	}

	sec := secrets.New()
	sec.SetPassword(password)
	sec.Set("url", urlStr)
	sec.Set("username", username)
	sec.Set("comment", comment)
	if u := pwrules.LookupChangeURL(hostname); u != "" {
		sec.Set("password-change-url", u)
	}
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(action.ExitEncrypt, err, "failed to set %q: %s", name, err)
	}

	return s.createPrintOrCopy(ctx, c, name, password, genPw)
}

// createPrintOrCopy will display the created password (or copy to clipboard)
func (s *creator) createPrintOrCopy(ctx context.Context, c *cli.Context, name, password string, genPw bool) error {
	if !genPw {
		return nil
	}

	if c.Bool("print") {
		fmt.Fprintf(
			out.Stdout,
			"The generated password for %s is:\n%s\n", name,
			color.YellowString(password),
		)
		return nil
	}

	if err := clipboard.CopyTo(ctx, name, []byte(password)); err != nil {
		return action.ExitError(action.ExitIO, err, "failed to copy to clipboard: %s", err)
	}
	return nil
}

// createPIN will walk through the numerical password (PIN) wizard
func (s *creator) createPIN(ctx context.Context, c *cli.Context) error {
	var (
		authority   = c.Args().Get(0)
		application = c.Args().Get(1)
		password    string
		comment     string
		store       = c.String("store")
		err         error
		genPw       bool
	)
	out.Printf(ctx, "=> Creating numerical PIN ...")
	authority, err = termio.AskForString(ctx, fmtfn(2, "1", "Authority"), authority)
	if err != nil {
		return err
	}
	if authority == "" {
		return action.ExitError(action.ExitUnknown, nil, "Authority must not be empty")
	}
	application, err = termio.AskForString(ctx, fmtfn(2, "2", "Entity"), application)
	if err != nil {
		return err
	}
	if application == "" {
		return action.ExitError(action.ExitUnknown, nil, "Application must not be empty")
	}
	genPw, err = termio.AskForBool(ctx, fmtfn(2, "3", "Generate PIN?"), false)
	if err != nil {
		return err
	}
	if genPw {
		password, err = s.createGeneratePIN(ctx)
		if err != nil {
			return err
		}
	} else {
		password, err = termio.AskForPassword(ctx, "PIN")
		if err != nil {
			return err
		}
	}
	comment, _ = termio.AskForString(ctx, fmtfn(2, "4", "Comments"), "")

	// select store
	if store == "" {
		store = cui.AskForStore(ctx, s.store)
	}

	// generate name, ask for override if already taken
	if store != "" {
		store += "/"
	}
	name := fmt.Sprintf("%spins/%s/%s", store, fsutil.CleanFilename(authority), fsutil.CleanFilename(application))
	if s.store.Exists(ctx, name) {
		name, err = termio.AskForString(ctx, fmtfn(2, "5", "Secret already exists, please choose another path"), name)
		if err != nil {
			return err
		}
	}
	sec := secrets.New()
	sec.SetPassword(password)
	sec.Set("application", application)
	sec.Set("comment", comment)
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(action.ExitEncrypt, err, "failed to set %q: %s", name, err)
	}

	return s.createPrintOrCopy(ctx, c, name, password, genPw)
}

// createGeneric will walk through the generic secret wizard
func (s *creator) createGeneric(ctx context.Context, c *cli.Context) error {
	var (
		shortname = c.Args().Get(0)
		password  string
		store     = c.String("store")
		err       error
		genPw     bool
	)
	out.Printf(ctx, "=> Creating generic secret ...")
	shortname, err = termio.AskForString(ctx, fmtfn(2, "1", "Name"), shortname)
	if err != nil {
		return err
	}
	if shortname == "" {
		return action.ExitError(action.ExitUnknown, nil, "Name must not be empty")
	}
	genPw, err = termio.AskForBool(ctx, fmtfn(2, "2", "Generate password?"), true)
	if err != nil {
		return err
	}
	if genPw {
		password, err = s.createGeneratePassword(ctx, "")
		if err != nil {
			return err
		}
	} else {
		password, err = termio.AskForPassword(ctx, shortname)
		if err != nil {
			return err
		}
	}

	// select store
	if store == "" {
		store = cui.AskForStore(ctx, s.store)
	}

	// generate name, ask for override if already taken
	if store != "" {
		store += "/"
	}
	name := fmt.Sprintf("%smisc/%s", store, fsutil.CleanFilename(shortname))
	if s.store.Exists(ctx, name) {
		name, err = termio.AskForString(ctx, "Secret already exists, please choose another path", name)
		if err != nil {
			return err
		}
	}
	sec := secrets.New()
	sec.SetPassword(password)
	out.Printf(ctx, fmtfn(2, "3", "Enter zero or more key value pairs for this secret:"))
	for {
		key, err := termio.AskForString(ctx, fmtfn(4, "a", "Name (enter to quit)"), "")
		if err != nil {
			return err
		}
		if key == "" {
			break
		}
		val, err := termio.AskForString(ctx, fmtfn(4, "b", "Value for Key '"+key+"'"), "")
		if err != nil {
			return err
		}
		sec.Set(key, val)
	}
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(action.ExitEncrypt, err, "failed to set %q: %s", name, err)
	}

	return s.createPrintOrCopy(ctx, c, name, password, genPw)
}

// createGeneratePasssword will walk through the password generation steps
func (s *creator) createGeneratePassword(ctx context.Context, hostname string) (string, error) {
	if _, found := pwrules.LookupRule(hostname); found {
		out.Printf(ctx, "Using password rules for %s ...", hostname)
		length, err := termio.AskForInt(ctx, fmtfn(4, "b", "How long?"), defaultLength)
		if err != nil {
			return "", err
		}
		return pwgen.NewCrypticForDomain(length, hostname).Password(), nil
	}
	xkcd, err := termio.AskForBool(ctx, fmtfn(4, "a", "Human-pronounceable passphrase? (see https://xkcd.com/936/)"), false)
	if err != nil {
		return "", err
	}
	if xkcd {
		length, err := termio.AskForInt(ctx, fmtfn(4, "b", "How many words?"), 4)
		if err != nil {
			return "", err
		}
		g := xkcdpwgen.NewGenerator()
		g.SetNumWords(length)
		g.SetDelimiter(" ")
		g.SetCapitalize(true)
		return string(g.GeneratePassword()), nil
	}

	length, err := termio.AskForInt(ctx, fmtfn(4, "b", "How long?"), defaultLength)
	if err != nil {
		return "", err
	}
	symbols, err := termio.AskForBool(ctx, fmtfn(4, "c", "Include symbols?"), false)
	if err != nil {
		return "", err
	}
	corp, err := termio.AskForBool(ctx, fmtfn(4, "d", "Strict rules?"), false)
	if err != nil {
		return "", err
	}
	if corp {
		return pwgen.GeneratePasswordWithAllClasses(length)
	}
	return pwgen.GeneratePassword(length, symbols), nil
}

// createGeneratePIN will walk through the PIN generation steps
func (s *creator) createGeneratePIN(ctx context.Context) (string, error) {
	length, err := termio.AskForInt(ctx, fmtfn(4, "a", "How long?"), 4)
	if err != nil {
		return "", err
	}
	return pwgen.GeneratePasswordCharset(length, "0123456789"), nil
}
