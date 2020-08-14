package pwgen

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

// GenerateExternal will invoke an external password generator,
// if set, and return it's output.
func GenerateExternal(pwlen int) (string, error) {
	c := os.Getenv("GOPASS_EXTERNAL_PWGEN")
	if c == "" {
		return "", fmt.Errorf("no external generator")
	}
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
	args = append(args, strconv.Itoa(pwlen))
	out, err := exec.Command(exe, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
