package action

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/qrcon"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()
	clip := c.Bool("clip")
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
		return err
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
