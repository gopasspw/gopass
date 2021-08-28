//go:build darwin
// +build darwin

package gitfs

// gitSSHCommand returns a SSH command instructing git to use SSH
// with persistent connections through a custom socket.
// See https://linux.die.net/man/5/ssh_config and
// https://git-scm.com/docs/git-config#Documentation/git-config.txt-coresshCommand
//
// Note: Setting GIT_SSH_COMMAND, possibly to an empty string, will take
// precedence over this setting.
//
// %C is a hash of %l%h%p%r and should avoid "path too long for unix domain socket"
// errors. On MacOS this doesn't always seem to work, so we're using a hardcoded
// /tmp instead.
func gitSSHCommand() string {
	return "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=/tmp/.ssh-%C"
}
