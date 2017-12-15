package action

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/qrcon"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	key := c.Args().Get(1)

	ctx = WithClip(ctx, c.Bool("clip"))
	ctx = WithForce(ctx, c.Bool("force"))
	ctx = WithPrintQR(ctx, c.Bool("qr"))
	ctx = WithPasswordOnly(ctx, c.Bool("password"))

	if err := s.show(ctx, c, name, key, true); err != nil {
		return exitError(ctx, ExitDecrypt, err, "%s", err)
	}
	return nil
}

func (s *Action) show(ctx context.Context, c *cli.Context, name, key string, recurse bool) error {
	if name == "" {
		return exitError(ctx, ExitUsage, nil, "Usage: %s show [name]", s.Name)
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

	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		if err != store.ErrNotFound || !recurse || !ctxutil.IsTerminal(ctx) {
			return exitError(ctx, ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
		}
		color.Yellow("Entry '%s' not found. Starting search...", name)
		if err := s.Find(ctx, c); err != nil {
			return exitError(ctx, ExitNotFound, err, "%s", err)
		}
		os.Exit(ExitNotFound)
	}

	var content string

	switch {
	case key != "":
		val, err := sec.Value(key)
		if err != nil {
			if errors.Cause(err) == store.ErrYAMLValueUnsupported {
				return exitError(ctx, ExitUnsupported, err, "Can not show nested key directly. Use 'gopass show %s'", name)
			}
			if errors.Cause(err) == store.ErrNotFound {
				return exitError(ctx, ExitNotFound, err, "Secret '%s' not found", name)
			}
			return exitError(ctx, ExitUnknown, err, "failed to retrieve key '%s' from '%s': %s", key, name, err)
		}
		if IsClip(ctx) {
			return s.copyToClipboard(ctx, name, []byte(val))
		}
		content = val
	case IsPrintQR(ctx):
		qr, err := qrcon.QRCode(sec.Password())
		if err != nil {
			return exitError(ctx, ExitUnknown, err, "failed to encode '%s' as QR: %s", name, err)
		}
		fmt.Println(qr)
		return nil
	case IsClip(ctx):
		return s.copyToClipboard(ctx, name, []byte(sec.Password()))
	default:
		switch {
		case IsPasswordOnly(ctx):
			content = sec.Password()
		case ctxutil.IsShowSafeContent(ctx) && !IsForce(ctx):
			content = sec.Body()
			if content == "" {
				return exitError(ctx, ExitNotFound, store.ErrNoBody, "no safe content to display, you can force display with show -f")
			}
		default:
			buf, err := sec.Bytes()
			if err != nil {
				return exitError(ctx, ExitUnknown, err, "failed to encode secret: %s", err)
			}
			content = string(buf)
		}
	}

	fmt.Print(color.YellowString(content))
	if ctxutil.IsTerminal(ctx) && !strings.HasSuffix(content, "\n") {
		fmt.Println("")
	}

	return nil
}
