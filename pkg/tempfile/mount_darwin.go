//go:build darwin
// +build darwin

package tempfile

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gopasspw/gopass/pkg/debug"
)

var shmDir = ""

func tempdirBase() string {
	return ""
}

func (t *File) mount(ctx context.Context) error {
	// create 32MB ramdisk
	cmd := exec.CommandContext(ctx, "hdid", "-drivekey", "system-image=yes", "-nomount", "ram://32768")
	cmd.Stderr = os.Stderr

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)
	cmdout, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create disk with hdid: %w", err)
	}

	debug.Log("Output: %s\n", cmdout)

	p := strings.Split(string(cmdout), " ")
	if len(p) < 1 {
		return fmt.Errorf("unhandeled hdid output: %s", string(cmdout))
	}
	t.dev = p[0]

	// create filesystem on ramdisk
	cmd = exec.CommandContext(ctx, "newfs_hfs", "-M", "700", t.dev)
	cmd.Stderr = os.Stderr

	if debug.IsEnabled() {
		cmd.Stdout = os.Stdout
	}

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make filesystem on %s: %w", t.dev, err)
	}

	// mount ramdisk
	cmd = exec.CommandContext(ctx, "diskutil", "mount", "nobrowse", "-mountOptions", "noatime", "-mountpoint", t.dir, t.dev)
	cmd.Stderr = os.Stderr
	if debug.IsEnabled() {
		cmd.Stdout = os.Stdout
	}

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount filesystem %s to %s: %w", t.dev, t.dir, err)
	}

	// Wait for the mount to settle. This is a hack.
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (t *File) unmount(ctx context.Context) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second

	return backoff.Retry(func() error {
		return t.tryUnmount(ctx)
	}, bo)
}

func (t *File) tryUnmount(ctx context.Context) error {
	if t.dir == "" || t.dev == "" {
		return nil
	}

	// unmount ramdisk
	cmd := exec.CommandContext(ctx, "diskutil", "unmountDisk", t.dev)
	cmd.Stderr = os.Stderr
	if debug.IsEnabled() {
		cmd.Stdout = os.Stdout
	}

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command '%+v': %w", cmd.Args, err)
	}

	// eject disk
	cmd = exec.CommandContext(ctx, "diskutil", "quiet", "eject", t.dev)
	cmd.Stderr = os.Stderr
	if debug.IsEnabled() {
		cmd.Stdout = os.Stdout
	}

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)

	return cmd.Run()
}
