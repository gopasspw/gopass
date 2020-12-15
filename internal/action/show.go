package action

import (
	"context"
	"fmt"
	"math/rand"
	"net/textproto"
	"path"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/internal/clipboard"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secret"
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
	ctx = WithClip(ctx, IsOnlyClip(ctx) || IsAlsoClip(ctx))
	return ctx
}

// Show the content of a secret file
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()

	ctx := showParseArgs(c)

	if key := c.Args().Get(1); key != "" {
		debug.Log("Setting key: %s", key)
		ctx = WithKey(ctx, key)
	}

	if err := s.show(ctx, c, name, true); err != nil {
		return ExitError(ExitDecrypt, err, "%s", err)
	}
	return nil
}

// show displays the given secret/key
func (s *Action) show(ctx context.Context, c *cli.Context, name string, recurse bool) error {
	if name == "" {
		return ExitError(ExitUsage, nil, "Usage: %s show [name]", s.Name)
	}

	if s.Store.IsDir(ctx, name) && !s.Store.Exists(ctx, name) {
		return s.List(c)
	}
	if s.Store.IsDir(ctx, name) && ctxutil.IsTerminal(ctx) {
		out.Cyan(ctx, "Warning: %s is a secret and a folder. Use 'gopass show %s' to display the secret and 'gopass list %s' to show the content of the folder", name, name, name)
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

// showHandleRevision displays a single revision
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

// showHandleOutput displays a secret
func (s *Action) showHandleOutput(ctx context.Context, name string, sec gopass.Secret) error {
	pw, body := s.showGetContent(ctx, sec)

	if pw == "" && body == "" {
		if ctxutil.IsShowSafeContent(ctx) && !ctxutil.IsForce(ctx) {
			out.Yellow(ctx, "Warning: safecontent=true. Use -f to display password, if any")
		}
		return ExitError(ExitNotFound, store.ErrEmptySecret, store.ErrEmptySecret.Error())
	}

	if IsPrintQR(ctx) && pw != "" {
		if err := s.showPrintQR(name, pw); err != nil {
			return err
		}
	}

	if IsClip(ctx) && pw != "" {
		if err := clipboard.CopyTo(ctx, name, []byte(pw)); err != nil {
			return err
		}
	}

	if body == "" {
		return nil
	}

	ctx = out.WithNewline(ctx, ctxutil.IsTerminal(ctx) && !strings.HasSuffix(body, "\n"))
	out.Yellow(ctx, body)
	return nil
}

func (s *Action) showGetContent(ctx context.Context, sec gopass.Secret) (string, string) {
	// YAML key
	if HasKey(ctx) {
		key := GetKey(ctx)
		val := sec.Values(key)
		if len(val) == 0 {
			return "", ""
		}
		body := val[0]
		if len(val) > 1 {
			body = strings.Join(val[1:], "\n")
		}
		debug.Log("Getting values for key %s: %s", key, len(val))
		return val[0], body
	}

	pw := sec.Get("password")
	fullBody := strings.TrimPrefix(string(sec.Bytes()), secret.Ident+"\n")

	// first line of the secret only
	if IsPrintQR(ctx) || IsOnlyClip(ctx) {
		return pw, ""
	}
	if IsPasswordOnly(ctx) {
		return pw, pw
	}

	// everything but the passwords and "unsafe" keys
	if ctxutil.IsShowSafeContent(ctx) && !ctxutil.IsForce(ctx) {
		var sb strings.Builder
		// Since we can have multiple entries per key, we need to make sure to keep track of the index to be able to preserve the ordering
		preserveOrder := make(map[string]int)
		for _, k := range sec.Keys() {
			currentIndex := preserveOrder[k]
			v := sec.Values(k)[currentIndex]
			preserveOrder[k]++
			sb.WriteString(k)
			sb.WriteString(": ")
			// check is this key should be obstructed
			if isUnsafeKey(k, sec) {
				debug.Log("obstructing unsafe key %s", k)
				sb.WriteString(randAsterisk())
			} else {
				sb.WriteString(v)
			}
			sb.WriteString("\n")
		}
		if len(sec.Keys()) > 0 && len(sec.GetBody()) > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(sec.GetBody())
		if IsAlsoClip(ctx) {
			return pw, sb.String()
		}
		return "", sb.String()
	}

	// everything (default)
	return sec.Get("password"), fullBody
}

func isUnsafeKey(key string, sec gopass.Secret) bool {
	if textproto.CanonicalMIMEHeaderKey(key) == "Password" {
		return true
	}
	uks := sec.Get("Unsafe-Keys")
	if uks == "" {
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
	return strings.Repeat("*", 5+rand.Intn(5))
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

// showHandleError handles errors retrieving secrets
func (s *Action) showHandleError(ctx context.Context, c *cli.Context, name string, recurse bool, err error) error {
	if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
		if IsClip(ctx) {
			_ = notify.Notify(ctx, "gopass - error", fmt.Sprintf("failed to retrieve secret '%s': %s", name, err))
		}
		return ExitError(ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
	}
	if newName := s.hasAliasDomain(ctx, name); newName != "" {
		return s.show(ctx, nil, newName, false)
	}
	if IsClip(ctx) {
		_ = notify.Notify(ctx, "gopass - warning", fmt.Sprintf("Entry '%s' not found. Starting search...", name))
	}
	out.Yellow(ctx, "Entry '%s' not found. Starting search...", name)
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
		return ExitError(ExitUnknown, err, "failed to encode '%s' as QR: %s", name, err)
	}
	fmt.Fprintln(stdout, qr)
	return nil
}
