//go:build darwin

package tempfile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gopasspw/gopass/pkg/debug"
)

var shmDir = ""

func tempdirBase() string {
	return ""
}

// ramdiskSpec is the size of the ramdisk we create, in 512 byte sectors.
// 32768 sectors * 512 byte = 16 MiB.
const ramdiskSpec = "ram://32768"

// hasNoDiskutilImage records that `diskutil image attach` is unavailable on
// this host, so we only pay for the failed exec once per process.
var hasNoDiskutilImage atomic.Bool

// attach creates a ramdisk and returns its device node.
//
// Recent macOS releases (confirmed on 27.0) deprecated the hdiutil disk image
// interface, which hdid(8) is a thin wrapper around. It still works, but prints
//
//	hdiutil: WARNING: 'hdiutil attach -nomount ...' is deprecated.
//	Please use 'diskutil image attach --noMount ...' instead.
//
// to stderr on every invocation, which leaks into the output of every gopass
// command that needs a secure tempdir, e.g. `gopass edit`. Prefer the
// replacement and fall back to hdid on releases that do not have it.
func attach(ctx context.Context) (string, error) {
	if hasNoDiskutilImage.Load() {
		return attachHdid(ctx)
	}

	dev, err := attachDiskutil(ctx)
	if err == nil {
		return dev, nil
	}

	debug.Log("diskutil image attach failed, falling back to hdid: %s", err)

	dev, err = attachHdid(ctx)
	if err != nil {
		return "", err
	}

	// hdid succeeded where diskutil did not, so this release predates the
	// `image` verb. Skip the probe for the rest of the process. A genuine
	// attach failure fails hdid too, and is not remembered.
	hasNoDiskutilImage.Store(true)

	return dev, nil
}

// attachDiskutil uses the `diskutil image attach` interface that supersedes
// hdiutil. It returns an error on releases that predate the `image` verb, which
// is why attach keeps the hdid path around.
//
// Stderr is captured rather than passed through: we may still fall back to
// hdid, and the user should not see the diagnostics of an attempt that did not
// end up mattering.
func attachDiskutil(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "diskutil", "image", "attach", "--noMount", "--nobrowse", ramdiskSpec)

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)

	cmdout, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			debug.Log("diskutil stderr: %s", ee.Stderr)
		}

		return "", fmt.Errorf("failed to create disk with diskutil: %w", err)
	}

	return parseDev(string(cmdout))
}

// attachHdid uses hdid(8), the only interface available on older releases.
func attachHdid(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "hdid", "-drivekey", "system-image=yes", "-nomount", ramdiskSpec)
	cmd.Stderr = os.Stderr

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)

	cmdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create disk with hdid: %w", err)
	}

	return parseDev(string(cmdout))
}

// parseDev extracts the device node from the output of an attach command.
// hdid pads its output with spaces and tabs, diskutil does not, so we can not
// simply cut at the first space.
func parseDev(out string) (string, error) {
	debug.Log("Output: %s", out)

	fields := strings.Fields(out)
	if len(fields) < 1 {
		return "", fmt.Errorf("no device node in attach output: %q", out)
	}

	return fields[0], nil
}

func (t *File) mount(ctx context.Context) error {
	// create 16MB ramdisk
	dev, err := attach(ctx)
	if err != nil {
		return err
	}

	t.dev = dev

	// create filesystem on ramdisk
	cmd := exec.CommandContext(ctx, "newfs_hfs", "-M", "700", t.dev)
	cmd.Stderr = os.Stderr

	if debug.IsEnabled() {
		cmd.Stdout = os.Stdout
	}

	debug.Log("CMD: %s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make filesystem on %s: %w", t.dev, err)
	}

	// mount ramdisk
	cmd = exec.CommandContext(ctx, "diskutil", "mount", "nobrowse", "-mountOptions", "rw,noatime", "-mountpoint", t.dir, t.dev)
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
