package action

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/pwgen"
	"github.com/justwatchcom/gopass/tpl"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/urfave/cli"
)

// Edit the content of a password file
func (s *Action) Edit(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("provide a secret name")
	}

	var content []byte
	var changed bool
	if s.Store.Exists(name) {
		var err error
		content, err = s.Store.Get(name)
		if err != nil {
			return fmt.Errorf("failed to decrypt %s: %v", name, err)
		}
	} else if tmpl, found := s.Store.LookupTemplate(name); found {
		changed = true
		// load template if it exists
		content = pwgen.GeneratePassword(defaultLength, false)
		if nc, err := tpl.Execute(string(tmpl), name, content, s.Store); err == nil {
			content = nc
		} else {
			fmt.Printf("failed to execute template: %s\n", err)
		}
	}

	nContent, err := s.editor(content)
	if err != nil {
		return err
	}

	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) && !changed {
		return nil
	}

	return s.Store.SetConfirm(name, nContent, fmt.Sprintf("Edited with %s", os.Getenv("EDITOR")), s.confirmRecipients)
}

func (s *Action) editor(content []byte) ([]byte, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "editor"
	}

	tmpfile, err := fsutil.TempFile("gopass-edit")
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create tmpfile %s: %s", editor, err)
	}
	defer func() {
		if err := tmpfile.Remove(); err != nil {
			color.Red("Failed to remove tempfile at %s: %s", tmpfile.Name(), err)
		}
	}()

	if _, err := tmpfile.Write(content); err != nil {
		return []byte{}, fmt.Errorf("failed to write tmpfile to start with %s %v: %s", editor, tmpfile.Name(), err)
	}
	if err := tmpfile.Close(); err != nil {
		return []byte{}, fmt.Errorf("failed to close tmpfile to start with %s %v: %s", editor, tmpfile.Name(), err)
	}

	cmdArgs, err := shellquote.Split(editor)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse EDITOR command `%s`", editor)
	}

	editor = cmdArgs[0]
	args := append(cmdArgs[1:], tmpfile.Name())
	cmd := exec.Command(editor, args...)
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

	// enforce unix line endings in the password store
	nContent = bytes.Replace(nContent, []byte("\r\n"), []byte("\n"), -1)
	nContent = bytes.Replace(nContent, []byte("\r"), []byte("\n"), -1)

	return nContent, nil
}
