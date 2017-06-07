package action

import (
	"bytes"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/password"
	"github.com/justwatchcom/gopass/qrcon"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()
	clip := c.Bool("clip")
	force := c.Bool("force")
	qr := c.Bool("qr")

	if name == "" {
		return fmt.Errorf("provide a secret name")
	}

	if s.Store.IsDir(name) {
		return s.List(c)
	}

	if clip || qr {
		content, err := s.Store.Get(name)
		if err != nil {
			return err
		}

		if qr {
			qr, err := qrcon.QRCode(string(content))
			if err != nil {
				return err
			}
			fmt.Println(qr)
			return nil
		}
		return s.copyToClipboard(name, content)
	}

	content, err := s.Store.Get(name)
	if err != nil {
		if err != password.ErrNotFound {
			return err
		}
		color.Yellow("Entry '%s' not found. Starting search...", name)
		return s.Find(c)
	}

	if s.Store.SafeContent && !force {
		lines := bytes.SplitN(content, []byte("\n"), 2)
		if len(lines) < 2 || len(bytes.TrimSpace(lines[1])) == 0 {
			return fmt.Errorf("no safe content to display, you can force display with show -f")
		}
		content = lines[1]
	}

	color.Yellow(string(content))

	return nil
}

func (s *Action) copyToClipboard(name string, content []byte) error {
	if err := clipboard.WriteAll(string(content)); err != nil {
		return err
	}

	if err := clearClipboard(content, s.Store.ClipTimeout); err != nil {
		return err
	}

	fmt.Printf("Copied %s to clipboard. Will clear in %d seconds.\n", color.YellowString(name), s.Store.ClipTimeout)
	return nil
}
