package action

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/qrcon"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	key := c.Args().Get(1)

	clip := c.Bool("clip")
	force := c.Bool("force")
	qr := c.Bool("qr")

	if err := s.show(ctx, c, name, key, clip, force, qr); err != nil {
		return s.exitError(ctx, ExitDecrypt, err, "%s", err)
	}
	return nil
}

func (s *Action) show(ctx context.Context, c *cli.Context, name, key string, clip, force, qr bool) error {
	if name == "" {
		return s.exitError(ctx, ExitUsage, nil, "Usage: %s show [name]", s.Name)
	}

	if s.Store.IsDir(ctx, name) {
		return s.List(ctx, c)
	}

	// auto-fallback to binary files with b64 suffix, if unique
	if !s.Store.Exists(ctx, name) && s.Store.Exists(ctx, name+BinarySuffix) {
		name += BinarySuffix
	}

	sec, err := s.Store.Get(ctx, name)
	if err != nil {
		if err != store.ErrNotFound {
			return s.exitError(ctx, ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
		}
		color.Yellow("Entry '%s' not found. Starting search...", name)
		if err := s.Find(ctx, c); err != nil {
			return s.exitError(ctx, ExitNotFound, err, "%s", err)
		}
		os.Exit(ExitNotFound)
	}

	var content string

	switch {
	case key != "":
		val, err := sec.Value(key)
		if err != nil {
			if errors.Cause(err) == store.ErrYAMLValueUnsupported {
				return s.exitError(ctx, ExitUnsupported, err, "Can not show nested key directly. Use 'gopass show %s'", name)
			}
			if errors.Cause(err) == store.ErrNotFound {
				return s.exitError(ctx, ExitNotFound, err, "Secret '%s' not found", name)
			}
			return s.exitError(ctx, ExitUnknown, err, "failed to retrieve key '%s' from '%s': %s", key, name, err)
		}
		if clip {
			return s.copyToClipboard(ctx, name, []byte(val))
		}
		content = val
	case qr:
		qr, err := qrcon.QRCode(sec.Password())
		if err != nil {
			return s.exitError(ctx, ExitUnknown, err, "failed to encode '%s' as QR: %s", name, err)
		}
		fmt.Println(qr)
		return nil
	case clip:
		return s.copyToClipboard(ctx, name, []byte(sec.Password()))
	default:
		if ctxutil.IsShowSafeContent(ctx) && !force {
			content = sec.Body()
			if content == "" {
				return s.exitError(ctx, ExitNotFound, store.ErrNoBody, "no safe content to display, you can force display with show -f")
			}
		} else {
			buf, err := sec.Bytes()
			if err != nil {
				return s.exitError(ctx, ExitUnknown, err, "failed to encode secret: %s", err)
			}
			content = string(buf)
		}
	}

	fmt.Print(color.YellowString(strings.TrimRight(string(content), "\r\n")))
	if ctxutil.IsTerminal(ctx) {
		fmt.Println("")
	}

	return nil
}

func (s *Action) copyToClipboard(ctx context.Context, name string, content []byte) error {
	if err := clipboard.WriteAll(string(content)); err != nil {
		return errors.Wrapf(err, "failed to write to clipboard")
	}

	if err := clearClipboard(ctx, content, ctxutil.GetClipTimeout(ctx)); err != nil {
		return errors.Wrapf(err, "failed to clear clipboard")
	}

	fmt.Printf("Copied %s to clipboard. Will clear in %d seconds.\n", color.YellowString(name), ctxutil.GetClipTimeout(ctx))
	return nil
}
