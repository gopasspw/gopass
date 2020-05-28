package action

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gopasspw/gopass/internal/clipboard"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/qrcon"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

const (
	// BinarySuffix is the suffix that is appended to binaries in the store
	BinarySuffix = ".b64"
)

func showParseArgs(c *cli.Context) context.Context {
	ctx := ctxutil.WithGlobalFlags(c)
	if c.IsSet("clip") {
		ctx = WithOnlyClip(ctx, c.Bool("clip"))
	}
	if c.IsSet("force") {
		ctx = ctxutil.WithForce(ctx, c.Bool("force"))
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
		ctx = WithKey(ctx, key)
	}

	if c.Bool("sync") {
		if err := s.sync(out.WithHidden(ctx, true), s.Store.MountPoint(name)); err != nil {
			out.Error(ctx, "Failed to sync %s: %s", name, err)
		}
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

	// auto-fallback to binary files with b64 suffix, if unique
	if !s.Store.Exists(ctx, name) && s.Store.Exists(ctx, name+BinarySuffix) {
		name += BinarySuffix
	}

	if HasRevision(ctx) {
		return s.showHandleRevision(ctx, c, name, GetRevision(ctx))
	}

	sec, ctx, err := s.Store.GetContext(ctx, name)
	if err != nil {
		return s.showHandleError(ctx, c, name, recurse, err)
	}

	return s.showHandleOutput(ctx, name, sec)
}

// showHandleRevision displays a single revision
func (s *Action) showHandleRevision(ctx context.Context, c *cli.Context, name, revision string) error {
	ctx, sec, err := s.Store.GetRevision(ctx, name, revision)
	if err != nil {
		return s.showHandleError(ctx, c, name, false, err)
	}

	return s.showHandleOutput(ctx, name, sec)
}

// showHandleOutput displays a secret
func (s *Action) showHandleOutput(ctx context.Context, name string, sec store.Secret) error {
	pw, body, err := s.showGetContent(ctx, name, sec)
	if err != nil {
		return err
	}

	if ctxutil.IsAutoClip(ctx) {
		ctx = WithClip(ctx, true)
	}

	if pw == "" && body == "" {
		return ExitError(ExitNotFound, store.ErrNoBody, store.ErrNoBody.Error())
	}

	if IsPrintQR(ctx) && pw != "" {
		return s.showPrintQR(name, pw)
	}

	if IsClip(ctx) && pw != "" && !ctxutil.IsForce(ctx) {
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

func (s *Action) showGetContent(ctx context.Context, name string, sec store.Secret) (string, string, error) {
	// YAML key
	if HasKey(ctx) {
		key := GetKey(ctx)
		val, err := sec.Value(key)
		if err != nil {
			return "", "", s.showHandleYAMLError(name, key, err)
		}
		return val, val, nil
	}

	// first line of the secret only
	if IsPrintQR(ctx) || IsOnlyClip(ctx) {
		return sec.Password(), "", nil
	}
	if IsPasswordOnly(ctx) {
		return sec.Password(), sec.Password(), nil
	}
	if ctxutil.IsAutoClip(ctx) && !ctxutil.IsForce(ctx) && !IsAlsoClip(ctx) {
		return sec.Password(), "", nil
	}

	// everything but the first line
	if ctxutil.IsShowSafeContent(ctx) && !ctxutil.IsForce(ctx) {
		if IsAlsoClip(ctx) {
			return sec.Password(), sec.Body(), nil
		}
		return "", sec.Body(), nil
	}

	// everything (default)
	buf, err := sec.Bytes()
	if err != nil {
		return "", "", ExitError(ExitUnknown, err, "failed to encode secret: %s", err)
	}
	return sec.Password(), string(buf), nil
}

// showHandleError handles errors retrieving secrets
func (s *Action) showHandleError(ctx context.Context, c *cli.Context, name string, recurse bool, err error) error {
	if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
		if IsClip(ctx) {
			_ = notify.Notify(ctx, "gopass - error", fmt.Sprintf("failed to retrieve secret '%s': %s", name, err))
		}
		return ExitError(ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
	}
	if IsClip(ctx) {
		_ = notify.Notify(ctx, "gopass - warning", fmt.Sprintf("Entry '%s' not found. Starting search...", name))
	}
	out.Yellow(ctx, "Entry '%s' not found. Starting search...", name)
	if err := s.Find(c); err != nil {
		if IsClip(ctx) {
			_ = notify.Notify(ctx, "gopass - error", fmt.Sprintf("%s", err))
		}
		return ExitError(ExitNotFound, err, "%s", err)
	}
	os.Exit(ExitNotFound)
	return nil
}

func (s *Action) showHandleYAMLError(name, key string, err error) error {
	if errors.Cause(err) == store.ErrYAMLValueUnsupported {
		return ExitError(ExitUnsupported, err, "Can not show nested key directly. Use 'gopass show %s'", name)
	}
	if errors.Cause(err) == store.ErrNotFound {
		return ExitError(ExitNotFound, err, "Secret '%s' not found", name)
	}
	return ExitError(ExitUnknown, err, "failed to retrieve key '%s' from '%s': %s", key, name, err)
}

func (s *Action) showPrintQR(name, pw string) error {
	qr, err := qrcon.QRCode(pw)
	if err != nil {
		return ExitError(ExitUnknown, err, "failed to encode '%s' as QR: %s", name, err)
	}
	fmt.Fprintln(stdout, qr)
	return nil
}
