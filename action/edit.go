package action

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/store/sub"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/pwgen"
	"github.com/justwatchcom/gopass/utils/tpl"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Edit the content of a password file
func (s *Action) Edit(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errors.Errorf("provide a secret name")
	}

	var content []byte
	var changed bool
	if s.Store.Exists(ctx, name) {
		sec, err := s.Store.Get(ctx, name)
		if err != nil {
			return errors.Errorf("failed to decrypt %s: %v", name, err)
		}
		content, err = sec.Bytes()
		if err != nil {
			return errors.Errorf("failed to decode %s: %v", name, err)
		}
	} else if tmpl, found := s.Store.LookupTemplate(ctx, name); found {
		changed = true
		// load template if it exists
		content = pwgen.GeneratePassword(defaultLength, false)
		if nc, err := tpl.Execute(ctx, string(tmpl), name, content, s.Store); err == nil {
			content = nc
		} else {
			fmt.Printf("failed to execute template: %s\n", err)
		}
	}

	nContent, err := s.editor(ctx, content)
	if err != nil {
		return s.exitError(ctx, ExitUnknown, err, "failed to invoke editor: %s", err)
	}

	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) && !changed {
		return nil
	}

	nSec, err := secret.Parse(nContent)
	if err != nil {
		fmt.Println(color.RedString("WARNING: Invalid YAML: %s", err))
	}

	if pw := nSec.Password(); pw != "" {
		printAuditResult(pw)
	}

	return s.Store.Set(sub.WithReason(ctx, fmt.Sprintf("Edited with %s", getEditor())), name, nSec)
}

func (s *Action) editor(ctx context.Context, content []byte) ([]byte, error) {
	editor := getEditor()

	tmpfile, err := fsutil.TempFile(ctx, "gopass-edit")
	if err != nil {
		return []byte{}, errors.Errorf("failed to create tmpfile %s: %s", editor, err)
	}
	defer func() {
		if err := tmpfile.Remove(ctx); err != nil {
			color.Red("Failed to remove tempfile at %s: %s", tmpfile.Name(), err)
		}
	}()

	if _, err := tmpfile.Write(content); err != nil {
		return []byte{}, errors.Errorf("failed to write tmpfile to start with %s %v: %s", editor, tmpfile.Name(), err)
	}
	if err := tmpfile.Close(); err != nil {
		return []byte{}, errors.Errorf("failed to close tmpfile to start with %s %v: %s", editor, tmpfile.Name(), err)
	}

	cmdArgs, err := shellquote.Split(editor)
	if err != nil {
		return []byte{}, errors.Errorf("failed to parse EDITOR command `%s`", editor)
	}

	editor = cmdArgs[0]
	args := append(cmdArgs[1:], tmpfile.Name())

	cmd := exec.Command(editor, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return []byte{}, errors.Errorf("failed to run %s with %s file", editor, tmpfile.Name())
	}

	nContent, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return []byte{}, errors.Errorf("failed to read from tmpfile: %v", err)
	}

	// enforce unix line endings in the password store
	nContent = bytes.Replace(nContent, []byte("\r\n"), []byte("\n"), -1)
	nContent = bytes.Replace(nContent, []byte("\r"), []byte("\n"), -1)

	return nContent, nil
}
