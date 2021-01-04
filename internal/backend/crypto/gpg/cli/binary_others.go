// +build !windows

package cli

import (
	"os/exec"
)

func detectBinary(name string) (string, error) {
	if name == "" {
		name = "gpg"
	}
	return exec.LookPath(name)
}
