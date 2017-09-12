// +build darwin

package fsutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/pkg/errors"
)

func tempdirBase() string {
	return ""
}

func (t *tempfile) mount(ctx context.Context) error {
	// create 16MB ramdisk
	cmd := exec.CommandContext(ctx, "hdid", "-drivekey", "system-image=yes", "-nomount", "ram://32768")
	cmd.Stderr = os.Stderr
	if ctxutil.IsDebug(ctx) {
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}

	out, err := cmd.Output()
	if err != nil {
		return errors.Errorf("Failed to create disk with hdid: %s", err)
	}

	if ctxutil.IsDebug(ctx) {
		fmt.Printf("[DEBUG] Output: %s\n", out)
	}

	p := strings.Split(string(out), " ")
	if len(p) < 1 {
		return errors.Errorf("Unhandeled hdid output: %s", string(out))
	}
	t.dev = p[0]

	// create filesystem on ramdisk
	cmd = exec.CommandContext(ctx, "newfs_hfs", "-M", "700", t.dev)
	cmd.Stderr = os.Stderr

	if ctxutil.IsDebug(ctx) {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	if err := cmd.Run(); err != nil {
		return errors.Errorf("Failed to make filesystem on %s: %s", t.dev, err)
	}

	// mount ramdisk
	cmd = exec.CommandContext(ctx, "mount", "-t", "hfs", "-o", "noatime", "-o", "nobrowse", t.dev, t.dir)
	cmd.Stderr = os.Stderr
	if t.dbg {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	if err := cmd.Run(); err != nil {
		return errors.Errorf("Failed to mount filesystem %s to %s: %s", t.dev, t.dir, err)
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (t *tempfile) unmount(ctx context.Context) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second
	return backoff.Retry(func() error {
		return t.tryUnmount(ctx)
	}, bo)
}

func (t *tempfile) tryUnmount(ctx context.Context) error {
	if t.dir == "" || t.dev == "" {
		return errors.Errorf("need dir and dev")
	}

	// unmount ramdisk
	cmd := exec.CommandContext(ctx, "diskutil", "unmountDisk", t.dev)
	cmd.Stderr = os.Stderr
	if t.dbg {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command '%+v'", cmd.Args)
	}

	// eject disk
	cmd = exec.CommandContext(ctx, "diskutil", "quiet", "eject", t.dev)
	cmd.Stderr = os.Stderr
	if t.dbg {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	return cmd.Run()
}
