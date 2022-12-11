package editor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
		args = append(args, vimOptions(resolveEditor(editor))...)
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
		debug.Log("Editor %s is not known to be vim compatible", editor)

		return []string{}
	}

	if !isVim(editor) {
		debug.Log("Editor %s is not known to be vim compatible", editor)

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

	args := []string{
		"-c",
		fmt.Sprintf("autocmd BufNewFile,BufRead %s setlocal noswapfile nobackup noundofile %s", path, viminfo),
	}
	args = append(args, "-i", "NONE") // disable viminfo
	args = append(args, "-n")         // disable swap

	return args
}

// isVim tries to identify the vi variant as vim compatible or not.
func isVim(editor string) bool {
	if editor == "neovim" {
		return true
	}

	cmd := exec.Command(editor, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		debug.Log("failed to check %s --version: %s", cmd.Path, err)

		return false
	}

	debug.Log("%s --version: %s", cmd.Path, string(out))

	return strings.Contains(string(out), "VIM - Vi IMproved")
}

// resolveEditor tries to resolve the final link destination of the editor name given
// and then extract the binary file name from the path. In practice the actual editor
// is often hidden behing several layers of indirection and we want to get an idea
// which options might work.
func resolveEditor(editor string) string {
	path, err := exec.LookPath(editor)
	if err != nil {
		debug.Log("failed to look up editor binary: %s", err)

		return editor
	}

	for {
		fi, err := os.Stat(path)
		if err != nil {
			debug.Log("failed to resolve %s: %s", path, err)

			return editor
		}

		if fi.Mode()&fs.ModeSymlink != fs.ModeSymlink {
			// not a symlink
			break
		}

		path, err = os.Readlink(path)
		if err != nil {
			debug.Log("failed to read link %s: %s", path, err)
		}
	}

	// return the binary name only
	return filepath.Base(path)
}
