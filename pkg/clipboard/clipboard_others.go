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
	"strconv"
	"syscall"

	ps "github.com/mitchellh/go-ps"
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

func walkFn(pid int, killFn func(int)) {
	// read the commandline for this process
	cmdline, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return
	}
	// compare the name of the binary and the first argument to avoid killing
	// any unrelated (gopass) processes
	args := bytes.Split(cmdline, []byte{0})
	if len(args) < 2 {
		return
	}
	// the commandline should start with "gopass"
	if string(args[0]) != os.Args[0] {
		return
	}
	// and have "unclip" as the first argument
	if string(args[1]) != "unclip" {
		return
	}

	killFn(pid)
}

func killProc(pid int) {
	// err should be always nil, but just to be sure
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	// we ignore this error as we're going to return nil anyway
	_ = proc.Kill()
}

// killPrecedessors will kill any previous "gopass unclip" invokations to avoid
// erasing the clipboard prematurely in case the the same content is copied to
// the clipboard repeately
func killPrecedessors() error {
	procs, err := ps.Processes()
	if err != nil {
		return err
	}
	for _, proc := range procs {
		walkFn(proc.Pid(), killProc)
	}
	return nil
}
