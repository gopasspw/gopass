package editor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/tempfile"
	shellquote "github.com/kballard/go-shellquote"
)

var (
	// Stdin is exported for tests.
	Stdin io.Reader = os.Stdin
	// Stdout is exported for tests.
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests.
	Stderr io.Writer = os.Stderr
)

// Invoke will start the given editor and return the content.
func Invoke(ctx context.Context, editor string, content []byte) ([]byte, error) {
	if !ctxutil.IsTerminal(ctx) {
		return nil, fmt.Errorf("need terminal")
	}

	tmpfile, err := tempfile.New(ctx, "gopass-edit")
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create tmpfile %s: %w", editor, err)
	}

	defer func() {
		if err := tmpfile.Remove(ctx); err != nil {
			color.Red("Failed to remove tempfile at %s: %s", tmpfile.Name(), err)
		}
	}()

	if _, err := tmpfile.Write(content); err != nil {
		return []byte{}, fmt.Errorf("failed to write tmpfile to start with %s %v: %w", editor, tmpfile.Name(), err)
	}

	if err := tmpfile.Close(); err != nil {
		return []byte{}, fmt.Errorf("failed to close tmpfile to start with %s %v: %w", editor, tmpfile.Name(), err)
	}

	args := make([]string, 0, 4)
	if runtime.GOOS != "windows" {
		cmdArgs, err := shellquote.Split(editor)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to parse EDITOR command `%s`", editor)
		}

		editor = cmdArgs[0]
		args = append(args, cmdArgs[1:]...)
		args = append(args, vimOptions(editor)...)
	}

	args = append(args, tmpfile.Name())

	cmd := exec.Command(editor, args...)
	cmd.Stdin = Stdin
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr

	if err := cmd.Run(); err != nil {
		debug.Log("cmd: %s %+v - error: %+v", cmd.Path, cmd.Args, err)

		return []byte{}, fmt.Errorf("failed to run %s with %s file: %w", editor, tmpfile.Name(), err)
	}

	nContent, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read from tmpfile: %w", err)
	}

	// enforce unix line endings in the password store.
	nContent = bytes.ReplaceAll(nContent, []byte("\r\n"), []byte("\n"))
	nContent = bytes.ReplaceAll(nContent, []byte("\r"), []byte("\n"))

	return nContent, nil
}

func vimOptions(editor string) []string {
	if editor != "vi" && editor != "vim" && editor != "neovim" {
		return []string{}
	}

	path := "/dev/shm/gopass*"
	if runtime.GOOS == "darwin" {
		path = "/private/**/gopass**"
	}
	viminfo := `viminfo=""`
	if editor == "neovim" {
		viminfo = `shada=""`
	}

	return []string{
		"-i", "NONE", // disable viminfo
		"-n", // disable swap
		"-c",
		fmt.Sprintf("autocmd BufNewFile,BufRead %s setlocal noswapfile nobackup noundofile %s", path, viminfo),
	}
}
