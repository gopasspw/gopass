// +build !windows

package clipboard

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// clear will spwan a copy of gopass that waits in a detached background
// process group until the timeout is expired. It will then compare the contents
// of the clipboard and erase it if it still contains the data gopass copied
// to it.
func clear(ctx context.Context, content []byte, timeout int) error {
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// kill any pending unclip processes
	_ = killPrecedessors()

	cmd := exec.Command(os.Args[0], "unclip", "--timeout", strconv.Itoa(timeout))
	// https://groups.google.com/d/msg/golang-nuts/shST-SDqIp4/za4oxEiVtI0J
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Env = append(os.Environ(), "GOPASS_UNCLIP_CHECKSUM="+hash)
	return cmd.Start()
}

// killPrecedessors will kill any previous "gopass unclip" invokations to avoid
// erasing the clipboard prematurely in case the the same content is copied to
// the clipboard repeately
func killPrecedessors() error {
	return filepath.Walk("/proc", func(path string, info os.FileInfo, err error) error {
		// ignore any errors as they are most likely due to insufficient privileges
		// (i.e. other users processes). We shouldn't get any errors for files we're
		// intrested in on a pseudo filesystem
		if err != nil {
			return nil
		}
		// only looking for paths that follow this pattern:
		// /proc/<pid>/status
		if strings.Count(path, "/") != 3 {
			return nil
		}
		if !strings.HasSuffix(path, "/status") {
			return nil
		}
		// extract the numeric pid
		pid, err := strconv.Atoi(path[6:strings.LastIndex(path, "/")])
		if err != nil {
			return nil
		}
		// read the commandline for this process
		cmdline, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if err != nil {
			return nil
		}
		// compare the name of the binary and the first argument to avoid killing
		// any unrelated (gopass) processes
		args := bytes.Split(cmdline, []byte{0})
		if len(args) < 2 {
			return nil
		}
		// the commandline should start with "gopass"
		if string(args[0]) != os.Args[0] {
			return nil
		}
		// and have "unclip" as the first argument
		if string(args[1]) != "unclip" {
			return nil
		}
		// err should be always nil, but just to be sure
		proc, err := os.FindProcess(pid)
		if err != nil {
			return nil
		}
		// we ignore this error as we're going to return nil anyway
		_ = proc.Kill()
		return nil
	})
}
