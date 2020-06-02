// +build darwin

package tempfile

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/debug"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

var shmDir = ""

func tempdirBase() string {
	return ""
}

func (t *File) mount(ctx context.Context) error {
	// create 16MB ramdisk
	cmd := exec.CommandContext(ctx, "hdid", "-drivekey", "system-image=yes", "-nomount", "ram://32768")
	cmd.Stderr = os.Stderr

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)
	cmdout, err := cmd.Output()
	if err != nil {
		return errors.Errorf("Failed to create disk with hdid: %s", err)
	}

	debug.Log("Output: %s\n", cmdout)

	p := strings.Split(string(cmdout), " ")
	if len(p) < 1 {
		return errors.Errorf("Unhandeled hdid output: %s", string(cmdout))
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
		return errors.Errorf("Failed to make filesystem on %s: %s", t.dev, err)
	}

	// mount ramdisk
	cmd = exec.CommandContext(ctx, "mount", "-t", "hfs", "-o", "noatime", "-o", "nobrowse", t.dev, t.dir)
	cmd.Stderr = os.Stderr
	if debug.IsEnabled() {
		cmd.Stdout = os.Stdout
	}

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return errors.Errorf("Failed to mount filesystem %s to %s: %s", t.dev, t.dir, err)
	}

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
		return errors.Wrapf(err, "failed to run command '%+v'", cmd.Args)
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
