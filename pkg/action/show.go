package action

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gopasspw/gopass/pkg/clipboard"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/notify"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/qrcon"
	"github.com/gopasspw/gopass/pkg/store"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

const (
	// BinarySuffix is the suffix that is appended to binaries in the store
	BinarySuffix = ".b64"
)

// Show the content of a secret file
func (s *Action) Show(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()

	ctx = s.Store.WithConfig(ctx, name)
	ctx = WithOnlyClip(ctx, c.Bool("clip"))
	ctx = WithClip(ctx, c.Bool("alsoclip"))
	ctx = WithForce(ctx, c.Bool("force"))
	ctx = WithPrintQR(ctx, c.Bool("qr"))
	ctx = WithPasswordOnly(ctx, c.Bool("password"))
	ctx = WithRevision(ctx, c.String("revision"))
	if key := c.Args().Get(1); key != "" {
		ctx = WithKey(ctx, key)
	}

	if c.Bool("sync") {
		if err := s.sync(out.WithHidden(ctx, true), c, s.Store.MountPoint(name)); err != nil {
			out.Error(ctx, "Failed to sync %s: %s", name, err)
		}
	}

	if err := s.show(ctx, c, name, true); err != nil {
		return ExitError(ctx, ExitDecrypt, err, "%s", err)
	}
	return nil
}

// show displays the given secret/key
func (s *Action) show(ctx context.Context, c *cli.Context, name string, recurse bool) error {
	if name == "" {
		return ExitError(ctx, ExitUsage, nil, "Usage: %s show [name]", s.Name)
	}

	if s.Store.IsDir(ctx, name) && !s.Store.Exists(ctx, name) {
		return s.List(ctx, c)
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
	sec, err := s.Store.GetRevision(ctx, name, revision)
	if err != nil {
		return s.showHandleError(ctx, c, name, false, err)
	}

	return s.showHandleOutput(ctx, name, sec)
}

// showHandleOutput displays a secret
func (s *Action) showHandleOutput(ctx context.Context, name string, sec store.Secret) error {
	var content string
	var key string
	if HasKey(ctx) {
		key = GetKey(ctx)
	}

	switch {
	case key != "":
		val, err := sec.Value(key)
		if err != nil {
			return s.showHandleYAMLError(ctx, name, key, err)
		}
		if IsClip(ctx) {
			if err := clipboard.CopyTo(ctx, name, []byte(val)); err != nil {
				return err
			}
			if IsOnlyClip(ctx) {
				return nil
			}
		}
		content = val
	case IsPrintQR(ctx):
		return s.showPrintQR(ctx, name, sec.Password())
	case IsClip(ctx):
		if err := clipboard.CopyTo(ctx, name, []byte(sec.Password())); err != nil {
			return err
		}
		if IsOnlyClip(ctx) {
			return nil
		}
		fallthrough
	default:
		switch {
		case IsPasswordOnly(ctx):
			content = sec.Password()
		case ctxutil.IsShowSafeContent(ctx) && !IsForce(ctx):
			content = sec.Body()
			if content == "" {
				if ctxutil.IsAutoClip(ctx) {
					out.Yellow(ctx, "Info: %s.", store.ErrNoBody.Error())
					out.Yellow(ctx, "Copying password instead.")
					return clipboard.CopyTo(ctx, name, []byte(sec.Password()))
				}
				return ExitError(ctx, ExitNotFound, store.ErrNoBody, store.ErrNoBody.Error())
			}
		default:
			buf, err := sec.Bytes()
			if err != nil {
				return ExitError(ctx, ExitUnknown, err, "failed to encode secret: %s", err)
			}
			content = string(buf)
		}
	}

	ctx = out.WithNewline(ctx, ctxutil.IsTerminal(ctx) && !strings.HasSuffix(content, "\n"))
	out.Yellow(ctx, content)
	return nil
}

// showHandleError handles errors retrieving secrets
func (s *Action) showHandleError(ctx context.Context, c *cli.Context, name string, recurse bool, err error) error {
	if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
		if IsClip(ctx) {
			_ = notify.Notify(ctx, "gopass - error", fmt.Sprintf("failed to retrieve secret '%s': %s", name, err))
		}
		return ExitError(ctx, ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
	}
	if IsClip(ctx) {
		_ = notify.Notify(ctx, "gopass - warning", fmt.Sprintf("Entry '%s' not found. Starting search...", name))
	}
	out.Yellow(ctx, "Entry '%s' not found. Starting search...", name)
	if err := s.Find(ctx, c); err != nil {
		if IsClip(ctx) {
			_ = notify.Notify(ctx, "gopass - error", fmt.Sprintf("%s", err))
		}
		return ExitError(ctx, ExitNotFound, err, "%s", err)
	}
	os.Exit(ExitNotFound)
	return nil
}

func (s *Action) showHandleYAMLError(ctx context.Context, name, key string, err error) error {
	if errors.Cause(err) == store.ErrYAMLValueUnsupported {
		return ExitError(ctx, ExitUnsupported, err, "Can not show nested key directly. Use 'gopass show %s'", name)
	}
	if errors.Cause(err) == store.ErrNotFound {
		return ExitError(ctx, ExitNotFound, err, "Secret '%s' not found", name)
	}
	return ExitError(ctx, ExitUnknown, err, "failed to retrieve key '%s' from '%s': %s", key, name, err)
}

func (s *Action) showPrintQR(ctx context.Context, name, pw string) error {
	qr, err := qrcon.QRCode(pw)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "failed to encode '%s' as QR: %s", name, err)
	}
	fmt.Fprintln(stdout, qr)
	return nil
}
