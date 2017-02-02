package action

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/justwatchcom/gopass/password"
	"github.com/urfave/cli"
)

// Insert a string as content to a secret file
func (s *Action) Insert(c *cli.Context) error {
	echo := c.Bool("echo")
	multiline := c.Bool("multiline")
	force := c.Bool("force")

	name := c.Args().Get(0)
	if name == "" {
		return fmt.Errorf("provide a secret name")
	}

	replacing, err := s.Store.Exists(name)
	if err != nil && err != password.ErrNotFound {
		return fmt.Errorf("failed to see if %s exists", name)
	}

	if !force { // don't check if it's force anyway
		if replacing && !askForConfirmation(fmt.Sprintf("An entry already exists for %s. Overwrite it?", name)) {
			return fmt.Errorf("not overwriting your current secret")
		}
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("Failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it
	if info.Mode()&os.ModeCharDevice == 0 {
		content := &bytes.Buffer{}

		if written, err := io.Copy(content, os.Stdin); err != nil {
			return fmt.Errorf("Failed to copy after %d bytes: %s", written, err)
		}

		return s.Store.SetConfirm(name, content.Bytes(), s.confirmRecipients)
	}

	// if multi-line input is requested start an editor
	if multiline {
		content, err := s.editor([]byte{})
		if err != nil {
			return err
		}
		return s.Store.SetConfirm(name, []byte(content), s.confirmRecipients)
	}

	// if echo mode is requested use a simple string input function
	var promptFn func(string) (string, error)
	if echo {
		promptFn = func(prompt string) (string, error) {
			return askForString(prompt, "")
		}
	}

	content, err := askForPassword(name, promptFn)
	if err != nil {
		return fmt.Errorf("failed to ask for password: %v", err)
	}

	return s.Store.SetConfirm(name, []byte(content), s.confirmRecipients)
}
