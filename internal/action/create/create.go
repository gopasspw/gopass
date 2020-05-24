package create

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/clipboard"
	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/martinhoefling/goxkcdpwgen/xkcdpwgen"
	"github.com/urfave/cli/v2"
)

const (
	defaultLength = 24
)

type storer interface {
	//Get(context.Context, string) (store.Secret, error)
	Set(context.Context, string, store.Byter) error
	Exists(context.Context, string) bool
	//Delete(context.Context, string) error
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
	acts = append(acts, cui.Action{Name: "AWS Secret Key", Fn: s.createAWS})
	acts = append(acts, cui.Action{Name: "GCP Service Account", Fn: s.createGCP})
	act, sel := cui.GetSelection(ctx, "Please select the type of secret you would like to create", acts.Selection())
	switch act {
	case "default":
		fallthrough
	case "show":
		return acts.Run(ctx, c, sel)
	default:
		return action.ExitError(ctx, action.ExitAborted, nil, "user aborted")
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
	out.Green(ctx, "=> Creating Website login")
	urlStr, err = termio.AskForString(ctx, fmtfn(2, "1", "URL"), urlStr)
	if err != nil {
		return err
	}
	// the hostname is used as part of the name
	hostname := extractHostname(urlStr)
	if hostname == "" {
		return action.ExitError(ctx, action.ExitUnknown, err, "Can not parse URL '%s'. Please use 'gopass edit' to manually create the secret", urlStr)
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
		password, err = s.createGeneratePassword(ctx)
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

	sec := secret.New(password, "")
	_ = sec.SetValue("url", urlStr)
	_ = sec.SetValue("username", username)
	_ = sec.SetValue("comment", comment)
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(ctx, action.ExitEncrypt, err, "failed to set '%s': %s", name, err)
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
		return action.ExitError(ctx, action.ExitIO, err, "failed to copy to clipboard: %s", err)
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
	out.Green(ctx, "=> Creating numerical PIN ...")
	authority, err = termio.AskForString(ctx, fmtfn(2, "1", "Authority"), authority)
	if err != nil {
		return err
	}
	if authority == "" {
		return action.ExitError(ctx, action.ExitUnknown, nil, "Authority must not be empty")
	}
	application, err = termio.AskForString(ctx, fmtfn(2, "2", "Entity"), application)
	if err != nil {
		return err
	}
	if application == "" {
		return action.ExitError(ctx, action.ExitUnknown, nil, "Application must not be empty")
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
	sec := secret.New(password, "")
	_ = sec.SetValue("application", application)
	_ = sec.SetValue("comment", comment)
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(ctx, action.ExitEncrypt, err, "failed to set '%s': %s", name, err)
	}

	return s.createPrintOrCopy(ctx, c, name, password, genPw)
}

// createAWS will walk through the AWS credential creation wizard
func (s *creator) createAWS(ctx context.Context, c *cli.Context) error {
	var (
		account   = c.Args().Get(0)
		username  = c.Args().Get(1)
		accesskey = c.Args().Get(2)
		secretkey string
		password  string
		region    string
		store     = c.String("store")
		err       error
	)
	out.Green(ctx, "=> Creating AWS credentials ...")
	account, err = termio.AskForString(ctx, fmtfn(2, "1", "AWS Account"), account)
	if err != nil {
		return err
	}
	if account == "" {
		return action.ExitError(ctx, action.ExitUnknown, nil, "Account must not be empty")
	}
	username, err = termio.AskForString(ctx, fmtfn(2, "2", "AWS IAM User"), username)
	if err != nil {
		return err
	}
	if username == "" {
		return action.ExitError(ctx, action.ExitUnknown, nil, "Username must not be empty")
	}
	password, err = termio.AskForString(ctx, fmtfn(2, "3", "AWS Account Password"), password)
	if err != nil {
		return err
	}
	accesskey, err = termio.AskForString(ctx, fmtfn(2, "4", "AWS_ACCESS_KEY_ID"), accesskey)
	if err != nil {
		return err
	}
	secretkey, err = termio.AskForPassword(ctx, "AWS_SECRET_ACCESS_KEY")
	if err != nil {
		return err
	}
	region, _ = termio.AskForString(ctx, fmtfn(2, "5", "AWS_DEFAULT_REGION"), "")

	// select store
	if store == "" {
		store = cui.AskForStore(ctx, s.store)
	}

	// generate name, ask for override if already taken
	if store != "" {
		store += "/"
	}
	name := fmt.Sprintf("%saws/iam/%s/%s", store, fsutil.CleanFilename(account), fsutil.CleanFilename(username))
	if s.store.Exists(ctx, name) {
		name, err = termio.AskForString(ctx, "Secret already exists, please choose another path", name)
		if err != nil {
			return err
		}
	}
	sec := secret.New(password, "")
	_ = sec.SetValue("account", account)
	_ = sec.SetValue("username", username)
	_ = sec.SetValue("accesskey", accesskey)
	_ = sec.SetValue("secretkey", secretkey)
	_ = sec.SetValue("region", region)
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(ctx, action.ExitEncrypt, err, "failed to set '%s': %s", name, err)
	}
	return nil
}

// createGCP will walk through the GCP credential creation wizard
func (s *creator) createGCP(ctx context.Context, c *cli.Context) error {
	var (
		project  string
		username string
		svcaccfn = c.Args().Get(0)
		store    = c.String("store")
		err      error
	)
	out.Green(ctx, "=> Creating GCP credentials ...")
	svcaccfn, err = termio.AskForString(ctx, fmtfn(2, "1", "Service Account JSON"), svcaccfn)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(svcaccfn)
	if err != nil {
		return err
	}
	username, project, err = extractGCPInfo(buf)
	if err != nil {
		return err
	}
	if username == "" {
		username, err = termio.AskForString(ctx, fmtfn(4, "a", "Account name"), "")
		if err != nil {
			return err
		}
	}
	if username == "" {
		return action.ExitError(ctx, action.ExitUnknown, nil, "Username must not be empty")
	}
	if project == "" {
		project, err = termio.AskForString(ctx, fmtfn(4, "b", "Project name"), "")
		if err != nil {
			return err
		}
	}
	if project == "" {
		return action.ExitError(ctx, action.ExitUnknown, nil, "Project must not be empty")
	}

	// select store
	if store == "" {
		store = cui.AskForStore(ctx, s.store)
	}

	// generate name, ask for override if already taken
	if store != "" {
		store += "/"
	}
	name := fmt.Sprintf("%sgcp/iam/%s/%s", store, fsutil.CleanFilename(project), fsutil.CleanFilename(username))
	if s.store.Exists(ctx, name) {
		name, err = termio.AskForString(ctx, fmtfn(2, "2", "Secret already exists, please choose another path"), name)
		if err != nil {
			return err
		}
	}
	sec := secret.New("", string(buf))
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(ctx, action.ExitEncrypt, err, "failed to set '%s': %s", name, err)
	}
	return nil
}

// extractGCPInfo will extract the GCP details from the given json blob
func extractGCPInfo(buf []byte) (string, string, error) {
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return "", "", err
	}
	p := strings.Split(m["client_email"], "@")
	if len(p) < 2 {
		return "", "", fmt.Errorf("client_email contains no email")
	}
	username := p[0]
	p = strings.Split(p[1], ".")
	if len(p) < 1 {
		return username, "", fmt.Errorf("hostname contains not enough separators")
	}
	return username, p[0], nil
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
	out.Green(ctx, "=> Creating generic secret ...")
	shortname, err = termio.AskForString(ctx, fmtfn(2, "1", "Name"), shortname)
	if err != nil {
		return err
	}
	if shortname == "" {
		return action.ExitError(ctx, action.ExitUnknown, nil, "Name must not be empty")
	}
	genPw, err = termio.AskForBool(ctx, fmtfn(2, "2", "Generate password?"), true)
	if err != nil {
		return err
	}
	if genPw {
		password, err = s.createGeneratePassword(ctx)
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
	sec := secret.New(password, "")
	out.Print(ctx, fmtfn(2, "3", "Enter zero or more key value pairs for this secret:"))
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
		_ = sec.SetValue(key, val)
	}
	if err := s.store.Set(ctxutil.WithCommitMessage(ctx, "Created new entry"), name, sec); err != nil {
		return action.ExitError(ctx, action.ExitEncrypt, err, "failed to set '%s': %s", name, err)
	}

	return s.createPrintOrCopy(ctx, c, name, password, genPw)
}

// createGeneratePasssword will walk through the password generation steps
func (s *creator) createGeneratePassword(ctx context.Context) (string, error) {
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
