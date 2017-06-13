package action

import (
	"fmt"

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

	if name == "" {
		return fmt.Errorf("provide a secret name")
	}

	if s.Store.IsDir(name) {
		return s.List(c)
	}

	var content []byte
	var err error

	switch {
	case key != "":
		content, err = s.Store.GetKey(name, key)
		if err != nil {
			return err
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
			return s.Find(c)
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

	fmt.Printf("Copied %s to clipboard. Will clear in %d seconds.\n", color.YellowString(name), s.Store.ClipTimeout)
	return nil
}
