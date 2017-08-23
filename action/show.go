package action

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/qrcon"
	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()
	key := c.Args().Get(1)

	clip := c.Bool("clip")
	force := c.Bool("force")
	qr := c.Bool("qr")

	if err := s.show(c, name, key, clip, force, qr); err != nil {
		return s.exitError(ExitDecrypt, err, "%s", err)
	}
	return nil
}

func (s *Action) show(c *cli.Context, name, key string, clip, force, qr bool) error {
	if name == "" {
		return s.exitError(ExitUsage, nil, "Usage: %s show [name]", s.Name)
	}

	if s.Store.IsDir(name) {
		return s.List(c)
	}

	// auto-fallback to binary files with b64 suffix, if unique
	if !s.Store.Exists(name) && s.Store.Exists(name+BinarySuffix) {
		name += BinarySuffix
	}

	var content []byte
	var err error

	switch {
	case key != "":
		content, err = s.Store.GetKey(name, key)
		if err != nil {
			if err == store.ErrYAMLValueUnsupported {
				return s.exitError(ExitUnsupported, err, "Can not show nested key directly. Use 'gopass show %s'", name)
			}
			return s.exitError(ExitUnknown, err, "failed to retrieve key '%s' from '%s': %s", key, name, err)
		}
		if clip {
			return s.copyToClipboard(name, content)
		}
	case qr:
		content, err = s.Store.GetFirstLine(name)
		if err != nil {
			return s.exitError(ExitDecrypt, err, "failed to retrieve secret '%s': %s", name, err)
		}
		qr, err := qrcon.QRCode(string(content))
		if err != nil {
			return s.exitError(ExitUnknown, err, "failed to encode '%s' as QR: %s", name, err)
		}
		fmt.Println(qr)
		return nil
	case clip:
		content, err = s.Store.GetFirstLine(name)
		if err != nil {
			return s.exitError(ExitDecrypt, err, "failed to retrieve secret '%s': %s", name, err)
		}
		return s.copyToClipboard(name, content)
	default:
		if s.Store.SafeContent() && !force {
			content, err = s.Store.GetBody(name)
		} else {
			content, err = s.Store.Get(name)
		}
		if err != nil {
			if err != store.ErrNotFound {
				return s.exitError(ExitUnknown, err, "failed to retrieve secret '%s': %s", name, err)
			}
			color.Yellow("Entry '%s' not found. Starting search...", name)
			if err := s.Find(c); err != nil {
				return s.exitError(ExitNotFound, err, "%s", err)
			}
			os.Exit(ExitNotFound)
		}
	}

	fmt.Println(color.YellowString(string(content)))

	return nil
}

func (s *Action) copyToClipboard(name string, content []byte) error {
	if err := clipboard.WriteAll(string(content)); err != nil {
		return errors.Wrapf(err, "failed to write to clipboard")
	}

	if err := clearClipboard(content, s.Store.ClipTimeout()); err != nil {
		return errors.Wrapf(err, "failed to clear clipboard")
	}

	fmt.Printf("Copied %s to clipboard. Will clear in %d seconds.\n", color.YellowString(name), s.Store.ClipTimeout())
	return nil
}
