package action

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/password"
	"github.com/urfave/cli"
)

// Edit the content of a password file
func (s *Action) Edit(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("provide a secret name")
	}

	exists, err := s.Store.Exists(name)
	if err != nil && err != password.ErrNotFound {
		return fmt.Errorf("failed to see if %s exists", name)
	}

	var content []byte
	if exists {
		content, err = s.Store.Get(name)
		if err != nil {
			return fmt.Errorf("failed to decrypt %s: %v", name, err)
		}
	}

	nContent, err := s.editor(content)
	if err != nil {
		return err
	}

	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) {
		return nil
	}

	return s.Store.SetConfirm(name, nContent, s.confirmRecipients)
}

func (s *Action) editor(content []byte) ([]byte, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return []byte{}, fmt.Errorf("failed to edit, please set $EDITOR")
	}

	tmpfile, err := ioutil.TempFile(fsutil.Tempdir(), "gopass-edit")
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create tmpfile to start with %s: %v", editor, tmpfile.Name())
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			log.Fatal(err)
		}
	}()

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return []byte{}, fmt.Errorf("failed to create tmpfile to start with %s: %v", editor, tmpfile.Name())
	}
	if err := tmpfile.Close(); err != nil {
		return []byte{}, fmt.Errorf("failed to create tmpfile to start with %s: %v", editor, tmpfile.Name())
	}

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return []byte{}, fmt.Errorf("failed to run %s with %s file", editor, tmpfile.Name())
	}

	nContent, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read from tmpfile: %v", err)
	}

	return nContent, nil
}
