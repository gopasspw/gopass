package hook

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/kballard/go-shellquote"
)

// Stderr is exported for tests.
var Stderr io.Writer = os.Stderr

func Invoke(ctx context.Context, hook string, hookArgs ...string) error {
	hCmd := strings.TrimSpace(config.String(ctx, hook))
	if hCmd == "" {
		return nil
	}
	if sv := os.Getenv("GOPASS_HOOK"); sv == "1" {
		debug.Log("GOPASS_HOOK=1, skipping reentrant hook execution")

		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	args := make([]string, 0, 4)
	if runtime.GOOS != "windows" {
		cmdArgs, err := shellquote.Split(hCmd)
		if err != nil {
			return fmt.Errorf("failed to parse hook command `%s`", hCmd)
		}

		hook = cmdArgs[0]
		args = append(args, cmdArgs[1:]...)
	}

	args = append(args, hookArgs...)

	cmd := exec.CommandContext(ctx, hook, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOPASS_HOOK=1")

	if err := cmd.Run(); err != nil {
		debug.Log("cmd: %s %+v - error: %+v", cmd.Path, cmd.Args, err)

		return fmt.Errorf("failed to run %s with %s %v", hook, args, err)
	}

	return nil
}
