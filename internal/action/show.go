package action

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/gopasspw/gopass/pkg/qrcon"
	"github.com/urfave/cli/v2"
)

func showParseArgs(c *cli.Context) context.Context {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("clip") {
		ctx = WithOnlyClip(ctx, c.Bool("clip"))
	}
	if c.IsSet("unsafe") {
		ctx = ctxutil.WithForce(ctx, c.Bool("unsafe"))
	}
	if c.IsSet("qr") {
		ctx = WithPrintQR(ctx, c.Bool("qr"))
	}
	if c.IsSet("password") {
		ctx = WithPasswordOnly(ctx, c.Bool("password"))
	}
	if c.IsSet("revision") {
		ctx = WithRevision(ctx, c.String("revision"))
	}
	if c.IsSet("alsoclip") {
		ctx = WithAlsoClip(ctx, c.Bool("alsoclip"))
	}
	if c.IsSet("noparsing") {
		ctx = ctxutil.WithShowParsing(ctx, !c.Bool("noparsing"))
	}
	ctx = WithClip(ctx, IsOnlyClip(ctx) || IsAlsoClip(ctx))
	return ctx
}

// Show the content of a secret file.
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()

	ctx := showParseArgs(c)

	if key := c.Args().Get(1); key != "" {
		debug.Log("Adding key to ctx: %s", key)
		ctx = WithKey(ctx, key)
	}

	if err := s.show(ctx, c, name, true); err != nil {
		return ExitError(ExitDecrypt, err, "%s", err)
	}
	return nil
}

// show displays the given secret/key.
func (s *Action) show(ctx context.Context, c *cli.Context, name string, recurse bool) error {
	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s show [name]", s.Name)
	}

	if s.Store.IsDir(ctx, name) && !s.Store.Exists(ctx, name) {
		return s.List(c)
	}
	if s.Store.IsDir(ctx, name) && ctxutil.IsTerminal(ctx) {
		out.Warningf(ctx, "%s is a secret and a folder. Use 'gopass show %s' to display the secret and 'gopass list %s' to show the content of the folder", name, name, name)
	}

	if HasRevision(ctx) {
		return s.showHandleRevision(ctx, c, name, GetRevision(ctx))
	}

	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		return s.showHandleError(ctx, c, name, recurse, err)
	}

	return s.showHandleOutput(ctx, name, sec)
}

// showHandleRevision displays a single revision.
func (s *Action) showHandleRevision(ctx context.Context, c *cli.Context, name, revision string) error {
	revision, err := s.parseRevision(ctx, name, revision)
	if err != nil {
		return ExitError(ExitUnknown, err, "Failed to get revisions: %s", err)
	}

	ctx, sec, err := s.Store.GetRevision(ctx, name, revision)
	if err != nil {
		return s.showHandleError(ctx, c, name, false, err)
	}

	return s.showHandleOutput(ctx, name, sec)
}

func (s *Action) parseRevision(ctx context.Context, name, revision string) (string, error) {
	debug.Log("Revision: %s", revision)
	if !strings.HasPrefix(revision, "-") {
		return revision, nil
	}

	revStr := strings.TrimPrefix(revision, "-")
	offset, err := strconv.Atoi(revStr)
	if err != nil {
		return "", err
	}

	debug.Log("Offset: %d", offset)
	revs, err := s.Store.ListRevisions(ctx, name)
	if err != nil {
		return "", err
	}

	if len(revs) < offset {
		debug.Log("Not enough revisions (%d)", len(revs))
		return revStr, nil
	}

	revision = revs[len(revs)-offset].Hash
	debug.Log("Found %s for offset %d", revision, offset)
	return revision, nil
}

// showHandleOutput displays a secret.
func (s *Action) showHandleOutput(ctx context.Context, name string, sec gopass.Secret) error {
	pw, body, err := s.showGetContent(ctx, sec)
	if err != nil {
		return err
	}

	if pw == "" && body == "" {
		if ctxutil.IsShowSafeContent(ctx) && !ctxutil.IsForce(ctx) {
			out.Warning(ctx, "safecontent=true. Use -f to display password, if any")
		}
		return ExitError(ExitNotFound, store.ErrEmptySecret, store.ErrEmptySecret.Error())
	}

	if IsPrintQR(ctx) && pw != "" {
		if err := s.showPrintQR(name, pw); err != nil {
			return err
		}
	}

	if IsClip(ctx) && pw != "" {
		if err := clipboard.CopyTo(ctx, name, []byte(pw), s.cfg.ClipTimeout); err != nil {
			return err
		}
	}

	if body == "" {
		return nil
	}

	ctx = out.WithNewline(ctx, ctxutil.IsTerminal(ctx))
	if ctxutil.IsTerminal(ctx) && !IsPasswordOnly(ctx) {
		header := fmt.Sprintf("Secret: %s\n", name)
		if HasKey(ctx) {
			header += fmt.Sprintf("Key: %s\n", GetKey(ctx))
		} else if ctxutil.IsShowParsing(ctx) {
			out.Warning(ctx, "Parsing is enabled. Use -n to disable.")
		}
		out.Print(ctx, header)
	}

	// output the actual secret, newlines are handled by ctx and Print.
	out.Print(ctx, out.Secret(body))

	return nil
}

func (s *Action) showGetContent(ctx context.Context, sec gopass.Secret) (string, string, error) {
	// YAML key.
	if HasKey(ctx) && ctxutil.IsShowParsing(ctx) {
		key := GetKey(ctx)
		values, found := sec.Values(key)
		if !found {
			return "", "", ExitError(ExitNotFound, store.ErrNoKey, store.ErrNoKey.Error())
		}
		val := strings.Join(values, "\n")
		return val, val, nil
	} else if HasKey(ctx) {
		out.Warning(ctx, "Parsing is disabled but a key was provided.")
		debug.Log("attempting to parse key %s with parsing disabled", GetKey(ctx))
	}

	pw := sec.Password()
	// fallback for old MIME secrets.
	fullBody := strings.TrimPrefix(string(sec.Bytes()), secrets.Ident+"\n")

	// first line of the secret only.
	if IsPrintQR(ctx) || IsOnlyClip(ctx) {
		return pw, "", nil
	}
	if IsPasswordOnly(ctx) {
		if pw == "" && fullBody != "" {
			return "", "", ExitError(ExitNotFound, store.ErrNoPassword, store.ErrNoPassword.Error())
		}
		return pw, pw, nil
	}

	// everything but the first line.
	if ctxutil.IsShowSafeContent(ctx) && !ctxutil.IsForce(ctx) {
		body := showSafeContent(ctx, sec)
		if IsAlsoClip(ctx) {
			return pw, body, nil
		}
		return "", body, nil
	}

	// everything (default).
	return sec.Password(), fullBody, nil
}

func showSafeContent(ctx context.Context, sec gopass.Secret) string {
	var sb strings.Builder
	for i, k := range sec.Keys() {
		sb.WriteString(k)
		sb.WriteString(": ")
		// check if this key should be obstructed.
		if isUnsafeKey(k, sec) {
			debug.Log("obstructing unsafe key %s", k)
			sb.WriteString(randAsterisk())
		} else {
			v, found := sec.Values(k)
			if !found {
				continue
			}
			sb.WriteString(strings.Join(v, "\n"+k+": "))
		}
		// we only add a final new line if the body is non-empty.
		if sec.Body() != "" || i < len(sec.Keys())-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString(sec.Body())
	return sb.String()
}

func isUnsafeKey(key string, sec gopass.Secret) bool {
	if strings.ToLower(key) == "password" {
		return true
	}

	uks, found := sec.Get("unsafe-keys")
	if !found || uks == "" {
		return false
	}

	for _, uk := range strings.Split(uks, ",") {
		uk = strings.TrimSpace(uk)
		if uk == "" {
			continue
		}
		if strings.EqualFold(uk, key) {
			return true
		}
	}

	return false
}

func randAsterisk() string {
	// we could also have a random number of asterisks but testing becomes painful.
	return strings.Repeat("*", 5)
}

func (s *Action) hasAliasDomain(ctx context.Context, name string) string {
	p := strings.Split(name, "/")
	for i := len(p) - 1; i > 0; i-- {
		d := p[i]
		for _, alias := range pwrules.LookupAliases(d) {
			sn := append(p[0:i], alias)
			sn = append(sn, p[i+1:]...)
			aliasName := strings.Join(sn, "/")
			if s.Store.Exists(ctx, aliasName) {
				return aliasName
			}
		}
		name = path.Dir(name)
	}
	return ""
}

// showHandleError handles errors retrieving secrets.
func (s *Action) showHandleError(ctx context.Context, c *cli.Context, name string, recurse bool, err error) error {
	if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
		if IsClip(ctx) {
			_ = notify.Notify(ctx, "gopass - error", fmt.Sprintf("failed to retrieve secret %q: %s", name, err))
		}
		return ExitError(ExitUnknown, err, "failed to retrieve secret %q: %s", name, err)
	}

	if newName := s.hasAliasDomain(ctx, name); newName != "" {
		return s.show(ctx, nil, newName, false)
	}

	if IsClip(ctx) {
		_ = notify.Notify(ctx, "gopass - warning", fmt.Sprintf("Entry %q not found. Starting search...", name))
	}

	out.Warningf(ctx, "Entry %q not found. Starting search...", name)
	c.Context = ctx
	if err := s.Find(c); err != nil {
		if IsClip(ctx) {
			_ = notify.Notify(ctx, "gopass - error", fmt.Sprintf("%s", err))
		}
		return ExitError(ExitNotFound, err, "%s", err)
	}
	return nil
}

func (s *Action) showPrintQR(name, pw string) error {
	qr, err := qrcon.QRCode(pw)
	if err != nil {
		return ExitError(ExitUnknown, err, "failed to encode %q as QR: %s", name, err)
	}
	fmt.Fprintln(stdout, qr)
	return nil
}
