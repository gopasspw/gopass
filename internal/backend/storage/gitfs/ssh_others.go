//go:build !windows && !darwin

package gitfs

import (
	"os"
	"runtime"

	"github.com/gopasspw/gopass/pkg/fsutil"
)

// gitSSHCommand returns a SSH command instructing git to use SSH
// with persistent connections through a custom socket.
// See https://linux.die.net/man/5/ssh_config and
// https://git-scm.com/docs/git-config#Documentation/git-config.txt-coresshCommand
//
// Note: Setting GIT_SSH_COMMAND, possibly to an empty string, will take
// precedence over this setting.
//
// %C is a hash of %l%h%p%r and should avoid "path too long for unix domain socket"
// errors. If you still encounter this error set TMPDIR to a short path, e.g. /tmp.
//
// On Android/Termux or when os.TempDir() is not an existing writable directory,
// persistent SSH connections are disabled to avoid persisting an unusable ControlPath.
func gitSSHCommand() string {
	return gitSSHCommandFor(runtime.GOOS, os.Getenv("TERMUX_VERSION"), os.TempDir())
}

func gitSSHCommandFor(goos, termuxVersion, tmpDir string) string {
	if goos == "android" || termuxVersion != "" {
		return ""
	}
	if !isWritableDir(tmpDir) {
		return ""
	}

	return "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=${TMPDIR:-/tmp}/.ssh-%C"
}

func isWritableDir(dir string) bool {
	if !fsutil.IsDir(dir) {
		return false
	}

	f, err := os.CreateTemp(dir, ".gopass-ssh-check-*")
	if err != nil {
		return false
	}

	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)

	return true
}
