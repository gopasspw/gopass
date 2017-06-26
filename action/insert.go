package action

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"
)

// Insert a string as content to a secret file
func (s *Action) Insert(c *cli.Context) error {
	echo := c.Bool("echo")
	multiline := c.Bool("multiline")
	force := c.Bool("force")

	confirm := s.confirmRecipients
	if force {
		confirm = nil
	}

	name := c.Args().Get(0)
	if name == "" {
		return fmt.Errorf("provide a secret name")
	}
	key := c.Args().Get(1)

	var content []byte
	var fromStdin bool

	info, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("Failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it
	if info.Mode()&os.ModeCharDevice == 0 {
		fromStdin = true
		buf := &bytes.Buffer{}

		if written, err := io.Copy(buf, os.Stdin); err != nil {
			return fmt.Errorf("Failed to copy after %d bytes: %s", written, err)
		}

		content = buf.Bytes()
	}

	// update to a single YAML entry
	if key != "" {
		if !fromStdin {
			pw, err := s.askForPassword(name+"/"+key, nil)
			if err != nil {
				return fmt.Errorf("failed to ask for password: %v", err)
			}
			content = []byte(pw)
		}
		return s.Store.SetKey(name, key, string(content))
	}

	if fromStdin {
		return s.Store.SetConfirm(name, content, "Read secret from STDIN", confirm)
	}

	if !force { // don't check if it's force anyway
		if s.Store.Exists(name) && !s.askForConfirmation(fmt.Sprintf("An entry already exists for %s. Overwrite it?", name)) {
			return fmt.Errorf("not overwriting your current secret")
		}
	}

	// if multi-line input is requested start an editor
	if multiline {
		content, err := s.editor([]byte{})
		if err != nil {
			return err
		}
		return s.Store.SetConfirm(name, []byte(content), fmt.Sprintf("Inserted user supplied password with %s", os.Getenv("EDITOR")), confirm)
	}

	// if echo mode is requested use a simple string input function
	var promptFn func(string) (string, error)
	if echo {
		promptFn = func(prompt string) (string, error) {
			return s.askForString(prompt, "")
		}
	}

	pw, err := s.askForPassword(name, promptFn)
	if err != nil {
		return fmt.Errorf("failed to ask for password: %v", err)
	}
	content = []byte(pw)

	return s.Store.SetConfirm(name, content, "Inserted user supplied password", confirm)
}
