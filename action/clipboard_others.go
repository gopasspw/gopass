// +build !windows

package action

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// clearClipboard will spwan a copy of gopass that waits in a detached background
// process group until the timeout is expired. It will then compare the contents
// of the clipboard and erase it if it still contains the data gopass copied
// to it.
func clearClipboard(ctx context.Context, content []byte, timeout int) error {
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "unclip", "--timeout", strconv.Itoa(timeout))
	// https://groups.google.com/d/msg/golang-nuts/shST-SDqIp4/za4oxEiVtI0J
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Env = append(os.Environ(), "GOPASS_UNCLIP_CHECKSUM="+hash)
	return cmd.Start()
}
