package pwgen

import (
	"fmt"
	"os/exec"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

func generateExternal(c string) (string, error) {
	cmdArgs, err := shellquote.Split(c)
	if err != nil {
		return "", err
	}
	if len(cmdArgs) < 1 {
		return "", fmt.Errorf("no command")
	}
	exe := cmdArgs[0]
	args := []string{}
	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}
	out, err := exec.Command(exe, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
