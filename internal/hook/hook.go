// Package hook provides a flexible hook system for gopass.
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
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/kballard/go-shellquote"
)

// Stderr is exported for tests.
var Stderr io.Writer = os.Stderr

type subStoreGetter interface {
	GetSubStore(string) (*leaf.Store, error)
	MountPoint(string) string
}

func InvokeRoot(ctx context.Context, hookName, secName string, s subStoreGetter, hookArgs ...string) error {
	sub, err := s.GetSubStore(s.MountPoint(secName))
	if err != nil {
		return err
	}

	return Invoke(ctx, hookName, sub.Storage().Path(), hookArgs...)
}

func Invoke(ctx context.Context, hook, dir string, hookArgs ...string) error {
	if true {
		// TODO(GH-2546) disabled until further discussion, cf. https://www.cvedetails.com/cve/CVE-2023-24055/

		return nil
	}

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

	if len(hook) > 2 && hook[:2] == "~/" {
		hook = appdir.UserHome() + hook[1:]
	}

	args = append(args, hookArgs...)

	cmd := exec.CommandContext(ctx, hook, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOPASS_HOOK=1")
	cmd.Dir = dir

	debug.Log("running hook %s with: %s %+v", hook, cmd.Path, cmd.Args)

	if err := cmd.Run(); err != nil {
		debug.Log("cmd: %s %+v - error: %+v", cmd.Path, cmd.Args, err)

		return fmt.Errorf("failed to run %s %v: %w", hook, args, err)
	}

	return nil
}
