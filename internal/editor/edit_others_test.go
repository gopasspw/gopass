//go:build !windows

package editor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestEditor(t *testing.T) {
	// necessary for setting up the env
	u := gptest.NewGUnitTester(t)
	assert.NotNil(t, u)

	ctx := config.NewContextInMemory()
	touch, err := exec.LookPath("touch")
	require.NoError(t, err, os.Getenv("PATH"))

	want := "foobar"
	out, err := Invoke(ctx, touch, []byte(want))
	require.NoError(t, err)
	if string(out) != want {
		t.Errorf("%q != %q", string(out), want)
	}
}

// runWithFlags executes a function inside a cli.Command action so that flags are parsed.
func runWithFlags(flags map[string]string, fn func(context.Context, *cli.Command)) {
	cliFlags := make([]cli.Flag, 0, len(flags))
	args := make([]string, 0, len(flags)+1)
	args = append(args, "test")
	for k, v := range flags {
		cliFlags = append(cliFlags, &cli.StringFlag{Name: k})
		args = append(args, "--"+k+"="+v)
	}
	cmd := &cli.Command{
		Flags: cliFlags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fn(ctx, cmd)

			return nil
		},
	}
	_ = cmd.Run(context.Background(), args)
}

func TestGetEditor(t *testing.T) {
	t.Setenv("EDITOR", "")
	td := t.TempDir()
	t.Setenv("GOPASS_HOMEDIR", td)

	t.Run("--editor=fooed", func(t *testing.T) {
		var got string
		runWithFlags(map[string]string{"editor": "fooed"}, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		assert.Equal(t, "fooed", got)
	})

	t.Run("/usr/bin/editor", func(t *testing.T) {
		var got string
		runWithFlags(nil, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		pathed, err := exec.LookPath("editor")
		if err == nil {
			assert.Equal(t, pathed, got)
		}
	})

	t.Run("EDITOR", func(t *testing.T) {
		t.Setenv("EDITOR", "fooenv")
		var got string
		runWithFlags(nil, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		assert.Equal(t, "fooenv", got)
	})

	t.Run("vi", func(t *testing.T) {
		t.Setenv("PATH", "/tmp")
		var got string
		runWithFlags(nil, func(ctx context.Context, cmd *cli.Command) {
			got = Path(ctx, cmd)
		})
		assert.Equal(t, "vi", got)
	})
}

func writeFakeEditor(t *testing.T, dir, name, script string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte("#!/bin/sh\n"+script+"\n"), 0o700))

	return p
}

func TestVimOptions(t *testing.T) {
	td := t.TempDir()

	nvimBin := writeFakeEditor(t, td, "nvim", `echo "NVIM v0.10.2"`)
	neovimBin := writeFakeEditor(t, td, "neovim", `echo "NVIM v0.10.2"`)
	vimBin := writeFakeEditor(t, td, "vim", `echo "VIM - Vi IMproved 9.1"`)
	vimBasicBin := writeFakeEditor(t, td, "vim.basic", `echo "VIM - Vi IMproved 9.1"`)
	viWrapperNvim := writeFakeEditor(t, td, "vi", `echo "NVIM v0.10.2"`)

	nviDir := t.TempDir()
	nviBin := writeFakeEditor(t, nviDir, "vi", `echo "Version 1.81.6 (nvi)"`)

	nanoBin := writeFakeEditor(t, td, "nano", `echo "should not be executed" >&2; exit 1`)

	// Symlink chains:
	//   editor -> alt-editor -> vim.basic
	//   vi-to-nvim -> nvim
	altLink := filepath.Join(td, "alt-editor")
	require.NoError(t, os.Symlink(vimBasicBin, altLink))
	editorLink := filepath.Join(td, "editor")
	require.NoError(t, os.Symlink(altLink, editorLink))

	viToNvim := filepath.Join(td, "vi-to-nvim")
	require.NoError(t, os.Symlink(nvimBin, viToNvim))

	wantNvim := []string{
		"-n",
		"-i", "NONE",
		"-c", "silent! set noswapfile nobackup nowritebackup | silent! set noundofile | silent! set shada=",
	}
	wantVim := []string{
		"-n",
		"-i", "NONE",
		"-c", "silent! set noswapfile nobackup nowritebackup | silent! set noundofile | silent! set viminfo=",
	}

	t.Run("nvim", func(t *testing.T) {
		assert.Equal(t, wantNvim, vimOptions(nvimBin))
	})

	t.Run("neovim", func(t *testing.T) {
		assert.Equal(t, wantNvim, vimOptions(neovimBin))
	})

	t.Run("vim", func(t *testing.T) {
		assert.Equal(t, wantVim, vimOptions(vimBin))
	})

	t.Run("symlink chain to vim.basic", func(t *testing.T) {
		assert.Equal(t, wantVim, vimOptions(editorLink))
	})

	t.Run("symlink to nvim", func(t *testing.T) {
		assert.Equal(t, wantNvim, vimOptions(viToNvim))
	})

	t.Run("vi wrapper calling nvim", func(t *testing.T) {
		assert.Equal(t, wantNvim, vimOptions(viWrapperNvim))
	})

	t.Run("nvi is not vim compatible", func(t *testing.T) {
		assert.Empty(t, vimOptions(nviBin))
	})

	t.Run("non-vim editor gets no options", func(t *testing.T) {
		assert.Empty(t, vimOptions(nanoBin))
	})
}

func TestInvokePassesNvimSecurityFlags(t *testing.T) {
	u := gptest.NewGUnitTester(t)
	assert.NotNil(t, u)

	td := t.TempDir()
	argsFile := filepath.Join(td, "args.txt")
	nvimBin := writeFakeEditor(t, td, "nvim", fmt.Sprintf(`printf '%%s\n' "$@" > %q`, argsFile))

	ctx := config.NewContextInMemory()
	out, err := Invoke(ctx, nvimBin, []byte("secret-value\n"))
	require.NoError(t, err)
	assert.Equal(t, "secret-value\n", string(out))

	rawArgs, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	gotArgs := strings.Split(strings.TrimRight(string(rawArgs), "\n"), "\n")
	require.GreaterOrEqual(t, len(gotArgs), 6)
	assert.Equal(t, "-n", gotArgs[0])
	assert.Equal(t, "-i", gotArgs[1])
	assert.Equal(t, "NONE", gotArgs[2])
	assert.Equal(t, "-c", gotArgs[3])
	assert.Contains(t, gotArgs[4], "noswapfile")
	assert.Contains(t, gotArgs[4], "nobackup")
	assert.Contains(t, gotArgs[4], "nowritebackup")
	assert.Contains(t, gotArgs[4], "noundofile")
	assert.Contains(t, gotArgs[4], "shada=")
}

func TestInvokeVimNoLeakFiles(t *testing.T) {
	vimPath, err := exec.LookPath("vim")
	if err != nil {
		t.Skip("vim not installed")
	}

	u := gptest.NewGUnitTester(t)
	assert.NotNil(t, u)

	td := t.TempDir()
	swapDir := filepath.Join(td, "swap")
	backupDir := filepath.Join(td, "backup")
	undoDir := filepath.Join(td, "undo")
	for _, d := range []string{swapDir, backupDir, undoDir} {
		require.NoError(t, os.MkdirAll(d, 0o700))
	}

	vimrc := filepath.Join(td, "vimrc")
	vimrcContent := fmt.Sprintf(
		"set swapfile\nset directory=%s//\nset backup\nset backupdir=%s//\nset writebackup\nset undofile\nset undodir=%s//\n",
		swapDir, backupDir, undoDir,
	)
	require.NoError(t, os.WriteFile(vimrc, []byte(vimrcContent), 0o600))

	// Suppress vim's non-TTY terminal warnings and escape sequences in test output.
	origStdout, origStderr := Stdout, Stderr
	Stdout, Stderr = io.Discard, io.Discard
	t.Cleanup(func() {
		Stdout, Stderr = origStdout, origStderr
	})

	// Configure vim to edit and save the file, then quit.
	editorCmd := fmt.Sprintf("%s -u %s -c 'normal oedited' -c wq", vimPath, vimrc)

	ctx := config.NewContextInMemory()
	got, err := Invoke(ctx, editorCmd, []byte("initial\n"))
	require.NoError(t, err)
	assert.Equal(t, "initial\nedited\n", string(got))

	for _, d := range []string{swapDir, backupDir, undoDir} {
		entries, err := os.ReadDir(d)
		require.NoError(t, err)
		assert.Empty(t, entries, "expected no leaked files in %s", d)
	}
}
