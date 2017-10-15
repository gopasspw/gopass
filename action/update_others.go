// +build !windows

package action

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

func (s *Action) updateGopass(ctx context.Context, version string, url string) error {
	exe, err := s.executable(ctx)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(url, "https") {
		return errors.Errorf("refusing non-https URL %s", url)
	}
	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	out.Debug(ctx, "Tempdir: %s", td)
	archive := filepath.Join(td, path.Base(url))
	if err := s.tryDownload(ctx, archive, url); err != nil {
		return err
	}
	binDst := exe + "_new"
	_ = os.Remove(binDst)
	if err := s.extract(ctx, archive, binDst); err != nil {
		return err
	}
	// launch rename script and exit
	out.Yellow(ctx, "Downloaded update. Exiting to install in place (%s) ...", exe)
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf(`sleep 1; echo ""; echo -n "Please wait ... "; sleep 2; mv "%s" "%s" && echo OK`, binDst, exe))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd.Start()
}

func (s *Action) isUpdateable(ctx context.Context) error {
	fn, err := s.executable(ctx)
	if err != nil {
		return err
	}
	out.Debug(ctx, "isUpdateable - File: %s", fn)
	// check file
	fi, err := os.Stat(fn)
	if err != nil {
		return err
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}
	if err := unix.Access(fn, unix.W_OK); err != nil {
		return err
	}
	// check dir
	fdir := filepath.Dir(fn)
	return unix.Access(fdir, unix.W_OK)
}

func (s *Action) executable(ctx context.Context) (string, error) {
	path, err := os.Executable()
	if err != nil {
		return path, err
	}
	path, err = filepath.EvalSymlinks(path)
	return path, err
}
