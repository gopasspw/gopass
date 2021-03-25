// +build !windows

package gitfs

import "os"

// gitSSHCommand returns a SSH command instructing git to use SSH
// with persistent connections through a custom socket.
// See https://linux.die.net/man/5/ssh_config and
// https://git-scm.com/docs/git-config#Documentation/git-config.txt-coresshCommand
//
// Note: Setting GIT_SSH_COMMAND, possibly to an empty string, will take
// precedence over this setting.
func gitSSHCommand() string {
	return "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=" + os.TempDir() + "/.gopass-ssh-${USER}-%r@%h:%p"
}
