// +build windows

package action

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// clearClipboard will spwan a copy of gopass that waits in a detached background
// process group until the timeout is expired. It will then compare the contents
// of the clipboard and erase it if it still contains the data gopass copied
// to it.
func clearClipboard(content []byte, timeout int) error {
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	cmd := exec.Command(os.Args[0], "unclip", "--timeout", strconv.Itoa(timeout))
	cmd.Env = append(os.Environ(), "GOPASS_UNCLIP_CHECKSUM="+hash)
	return cmd.Start()
}
