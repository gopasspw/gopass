package editor

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/tempfile"

	"github.com/fatih/color"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
)

var (
	// Stdin is exported for tests
	Stdin io.Reader = os.Stdin
	// Stdout is exported for tests
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests
	Stderr io.Writer = os.Stderr
)

// Invoke will start the given editor and return the content
func Invoke(ctx context.Context, editor string, content []byte) ([]byte, error) {
	if !ctxutil.IsTerminal(ctx) {
		return nil, errors.New("need terminal")
	}

	tmpfile, err := tempfile.New(ctx, "gopass-edit")
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

	var args []string
	if runtime.GOOS != "windows" {
		cmdArgs, err := shellquote.Split(editor)
		if err != nil {
			return []byte{}, errors.Errorf("failed to parse EDITOR command `%s`", editor)
		}

		editor = cmdArgs[0]
		args = append(cmdArgs[1:], tmpfile.Name())
	} else {
		args = []string{tmpfile.Name()}
	}

	cmd := exec.Command(editor, args...)
	cmd.Stdin = Stdin
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr

	if err := cmd.Run(); err != nil {
		out.Debug(ctx, "editor - cmd: %s %+v - error: %+v", cmd.Path, cmd.Args, err)
		return []byte{}, errors.Errorf("failed to run %s with %s file: %s", editor, tmpfile.Name(), err)
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
