package action

import (
	"bytes"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

// Show the content of a secret file
func (s *Action) Show(c *cli.Context) error {
	name := c.Args().First()
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

	if c.Bool("clip") {
		return s.copyToClipboard(name, content)
	}

	color.Yellow(string(content))

	return nil
}

func (s *Action) copyToClipboard(name string, content []byte) error {
	content = bytes.TrimSpace(content)

	// only copy the first line to the clipboard
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) < 1 {
		return fmt.Errorf("no content that can be copied to the clipboard")
	}
	line := lines[0]

	if err := clipboard.WriteAll(string(line)); err != nil {
		return err
	}
	if err := clearClipboard(line, s.Store.ClipTimeout); err != nil {
		return err
	}
	fmt.Printf("Copied %s to clipboard. Will clear in %d seconds.\n", color.YellowString(name), s.Store.ClipTimeout)
	return nil
}
