package action

import (
	"bytes"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/qrcon"
	"github.com/smallfish/simpleyaml"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()
	clip := c.Bool("clip")
	force := c.Bool("force")
	qr := c.Bool("qr")
	key := c.String("key")

	if name == "" {
		return fmt.Errorf("provide a secret name")
	}

	if s.Store.IsDir(name) {
		return s.List(c)
	}

	content, err := s.Store.Get(name)
	if err != nil {
		return err
	}

	// if we only want to display safe contnt or if we want parse the secret as yaml strip the first line
	if (s.Store.SafeContent && !force) || key != "" {
		content, err = s.Store.Metadata(name)
		if err != nil {
			return err
		}
	} else if clip || qr {
		content, err = s.Store.First(name)
		if err != nil {
			return err
		}
	}

	if key != "" {
		yaml, err := simpleyaml.NewYaml(content)
		if err != nil {
			return fmt.Errorf("failed to load secret as yaml")
		} else {
			value, err := yaml.Get(key).String()
			if err != nil {
				keys, err := yaml.GetMapKeys()
				if err == nil {
					return fmt.Errorf("%s not available. Available keys are: %s\n", key, keys)
				}
			} else {
				content = []byte(value)
			}
		}
	}

	if clip || qr {
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
