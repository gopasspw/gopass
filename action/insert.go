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
		return s.exitError(ExitNoName, nil, "Usage: %s insert name", s.Name)
	}

	key := c.Args().Get(1)

	var content []byte
	var fromStdin bool

	info, err := os.Stdin.Stat()
	if err != nil {
		return s.exitError(ExitIO, err, "failed to stat stdin: %s", err)
	}

	// if content is piped to stdin, read and save it
	if info.Mode()&os.ModeCharDevice == 0 {
		fromStdin = true
		buf := &bytes.Buffer{}

		if written, err := io.Copy(buf, os.Stdin); err != nil {
			return s.exitError(ExitIO, err, "failed to copy after %d bytes: %s", written, err)
		}

		content = buf.Bytes()
	}

	// update to a single YAML entry
	if key != "" {
		if !fromStdin {
			pw, err := s.askForString(name+":"+key, "")
			if err != nil {
				return s.exitError(ExitIO, err, "failed to ask for user input: %s", err)
			}
			content = []byte(pw)
		}

		if err := s.Store.SetKey(name, key, string(content)); err != nil {
			return s.exitError(ExitEncrypt, err, "failed to set key '%s' of '%s': %s", key, name, err)
		}
		return nil
	}

	if fromStdin {
		if err := s.Store.SetConfirm(name, content, "Read secret from STDIN", confirm); err != nil {
			return s.exitError(ExitEncrypt, err, "failed to set '%s': %s", name, err)
		}
		return nil
	}

	if !force { // don't check if it's force anyway
		if s.Store.Exists(name) && !s.askForConfirmation(fmt.Sprintf("An entry already exists for %s. Overwrite it?", name)) {
			return s.exitError(ExitAborted, nil, "not overwriting your current secret")
		}
	}

	// if multi-line input is requested start an editor
	if multiline {
		content, err := s.editor([]byte{})
		if err != nil {
			return s.exitError(ExitUnknown, err, "failed to start editor: %s", err)
		}
		if err := s.Store.SetConfirm(name, content, fmt.Sprintf("Inserted user supplied password with %s", os.Getenv("EDITOR")), confirm); err != nil {
			return s.exitError(ExitEncrypt, err, "failed to store secret '%s': %s", name, err)
		}
		return nil
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
		return s.exitError(ExitIO, err, "failed to ask for password: %s", err)
	}

	printAuditResult(pw)
	content = []byte(pw)

	if err := s.Store.SetConfirm(name, content, "Inserted user supplied password", confirm); err != nil {
		return s.exitError(ExitEncrypt, err, "failed to write secret '%s': %s", name, err)
	}
	return nil
}
