//go:build windows
// +build windows

package gitconfig

import (
	"os/exec"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/debug"
)

var systemConfig string

func init() {
	gitPath, err := exec.LookPath("git.exe")
	if err != nil {
		debug.Log("git not found in PATH. Can not determine system config location")

		return
	}

	// gitPath is something like C:\Program Files\Git\cmd\git.exe
	// we need to strip the last two components to get the base path
	// and then append etc/gitconfig.
	systemConfig = filepath.Join(filepath.Dir(filepath.Dir(gitPath)), "etc", "gitconfig")
}
