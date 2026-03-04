//go:build !windows && !darwin

package gitfs

import "os"

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
func gitSSHCommand() string {
	return "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=" + os.TempDir() + "/.ssh-%C"
}
