// Package editor provides a simple wrapper around the EDITOR environment variable.
package editor

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

	args := make([]string, 0, 5)
	if runtime.GOOS != "windows" {
		cmdArgs, err := shellquote.Split(editor)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to parse EDITOR command `%s`", editor)
		}

		if len(cmdArgs) == 0 {
			return []byte{}, fmt.Errorf("empty EDITOR command")
		}

		editor = cmdArgs[0]
		// Validate the editor binary actually exists before attempting to run it.
		if _, err := exec.LookPath(editor); err != nil {
			return []byte{}, fmt.Errorf("editor binary %q not found in PATH: %w", editor, err)
		}

		args = append(args, vimOptions(editor)...)
		args = append(args, cmdArgs[1:]...)
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

type editorKind int

const (
	editorUnknown editorKind = iota
	editorVim
	editorNeovim
)

func vimOptions(editor string) []string {
	kind := detectEditorKind(editor)
	if kind == editorUnknown {
		return []string{}
	}

	stateOpt := "viminfo="
	if kind == editorNeovim {
		stateOpt = "shada="
	}

	return []string{
		"-n",         // disable swap file
		"-i", "NONE", // disable viminfo / shada
		"-c", fmt.Sprintf("silent! set noswapfile nobackup nowritebackup | silent! set noundofile | silent! set %s", stateOpt),
	}
}

// detectEditorKind tries to identify whether the given editor is Vim or Neovim.
func detectEditorKind(editor string) editorKind {
	resolved := resolveEditorPath(editor)
	base := filepath.Base(resolved)

	switch {
	case base == "nvim" || base == "neovim":
		return editorNeovim
	case base == "vi" || base == "vim" || base == "gvim" || base == "mvim" || strings.HasPrefix(base, "vim."):
		return probeVimVersion(resolved)
	default:
		debug.Log("Editor %s (%s) is not known to be vim compatible", editor, base)

		return editorUnknown
	}
}

func probeVimVersion(editorPath string) editorKind {
	cmd := exec.Command(editorPath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		debug.Log("failed to check %s --version: %s", cmd.Path, err)

		return editorUnknown
	}

	version := string(out)
	debug.Log("%s --version: %s", cmd.Path, version)

	if strings.Contains(version, "NVIM ") {
		return editorNeovim
	}
	if strings.Contains(version, "VIM - Vi IMproved") {
		return editorVim
	}

	return editorUnknown
}

// resolveEditorPath tries to resolve the final symlink destination of the editor binary.
// In practice the actual editor is often hidden behind several layers of indirection
// (e.g. /usr/bin/editor -> /etc/alternatives/editor -> /usr/bin/vim.basic or nvim)
// and we want to know which options are safe and effective to pass.
func resolveEditorPath(editor string) string {
	path, err := exec.LookPath(editor)
	if err != nil {
		debug.Log("failed to look up editor binary: %s", err)

		return editor
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		debug.Log("failed to resolve symlinks for %s: %s", path, err)

		return path
	}

	return resolved
}
