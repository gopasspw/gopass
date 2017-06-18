// +build darwin

package fsutil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
)

func tempdirBase() string {
	return ""
}

func (t *tempfile) mount() error {
	// create 16MB ramdisk
	cmd := exec.Command("hdid", "-drivekey", "system-image=yes", "-nomount", "ram://32768")
	cmd.Stderr = os.Stderr
	if t.dbg {
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to create disk with hdid: %s", err)
	}
	if t.dbg {
		fmt.Printf("[DEBUG] Output: %s\n", out)
	}
	p := strings.Split(string(out), " ")
	if len(p) < 1 {
		return fmt.Errorf("Unhandeled hdid output: %s", string(out))
	}
	t.dev = p[0]

	// create filesystem on ramdisk
	cmd = exec.Command("newfs_hfs", "-M", "700", t.dev)
	cmd.Stderr = os.Stderr
	if t.dbg {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to make filesystem on %s: %s", t.dev, err)
	}

	// mount ramdisk
	cmd = exec.Command("mount", "-t", "hfs", "-o", "noatime", "-o", "nobrowse", t.dev, t.dir)
	cmd.Stderr = os.Stderr
	if t.dbg {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to mount filesystem %s to %s: %s", t.dev, t.dir, err)
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (t *tempfile) unmount() error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second
	return backoff.Retry(t.tryUnmount, bo)
}

func (t *tempfile) tryUnmount() error {
	if t.dir == "" || t.dev == "" {
		return fmt.Errorf("need dir and dev")
	}
	// unmount ramdisk
	cmd := exec.Command("diskutil", "unmountDisk", t.dev)
	cmd.Stderr = os.Stderr
	if t.dbg {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// eject disk
	cmd = exec.Command("diskutil", "quiet", "eject", t.dev)
	cmd.Stderr = os.Stderr
	if t.dbg {
		cmd.Stdout = os.Stdout
		fmt.Printf("[DEBUG] CMD: %s %+v\n", cmd.Path, cmd.Args)
	}
	return cmd.Run()
}
