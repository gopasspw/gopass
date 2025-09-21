package pwgen

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

var (
	// ErrNoExternal is returned when no external generator is set.
	ErrNoExternal = fmt.Errorf("no external generator")
	// ErrNoCommand is returned when no command is set.
	ErrNoCommand = fmt.Errorf("no command")
)

// GenerateExternal will invoke an external password generator,
// if set, and return its output.
// The external generator is configured via the GOPASS_EXTERNAL_PWGEN environment variable.
func GenerateExternal(pwlen int) (string, error) {
	c := os.Getenv("GOPASS_EXTERNAL_PWGEN")
	if c == "" {
		return "", ErrNoExternal
	}

	cmdArgs, err := shellquote.Split(c)
	if err != nil {
		return "", fmt.Errorf("failed to split %s: %w", c, err)
	}

	if len(cmdArgs) < 1 {
		return "", ErrNoCommand
	}

	exe := cmdArgs[0]
	args := []string{}

	if len(cmdArgs) > 1 {
		args = cmdArgs[1:]
	}

	args = append(args, strconv.Itoa(pwlen))

	out, err := exec.Command(exe, args...).Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute %s %v: %w", exe, args, err)
	}

	return strings.TrimSpace(string(out)), nil
}
