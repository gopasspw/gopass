package action

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/qrcon"
	"github.com/justwatchcom/gopass/store"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()
	key := c.Args().Get(1)

	clip := c.Bool("clip")
	force := c.Bool("force")
	qr := c.Bool("qr")

	return s.show(c, name, key, clip, force, qr)
}

func (s *Action) show(c *cli.Context, name, key string, clip, force, qr bool) error {
	if name == "" {
		return fmt.Errorf("provide a secret name")
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
				return fmt.Errorf("Can not show nested key directly. Use 'gopass show %s'", name)
			}
			return err
		}
		if clip {
			return s.copyToClipboard(name, content)
		}
	case qr:
		content, err = s.Store.GetFirstLine(name)
		if err != nil {
			return err
		}
		qr, err := qrcon.QRCode(string(content))
		if err != nil {
			return err
		}
		fmt.Println(qr)
		return nil
	case clip:
		content, err = s.Store.GetFirstLine(name)
		if err != nil {
			return err
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
				return err
			}
			color.Yellow("Entry '%s' not found. Starting search...", name)
			if err := s.Find(c); err != nil {
				return err
			}
			os.Exit(1)
		}
	}

	color.Yellow(string(content))

	return nil
}

func (s *Action) copyToClipboard(name string, content []byte) error {
	if err := clipboard.WriteAll(string(content)); err != nil {
		return err
	}

	if err := clearClipboard(content, s.Store.ClipTimeout()); err != nil {
		return err
	}

	fmt.Printf("Copied %s to clipboard. Will clear in %d seconds.\n", color.YellowString(name), s.Store.ClipTimeout())
	return nil
}
