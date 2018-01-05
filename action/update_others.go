// +build !windows

package action

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
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

var (
	updateMoveAfterQuit = true
)

func (s *Action) updateGopass(ctx context.Context, version string, urlStr string) error {
	exe, err := s.executable(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to detect executable location")
	}
	out.Debug(ctx, "Excuteable is at '%s'", exe)

	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.Wrapf(err, "failed to parse URL")
	}

	if err := updateCheckHost(ctx, u); err != nil {
		return err
	}

	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		return errors.Wrapf(err, "failed to create tempdir")
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	out.Debug(ctx, "Tempdir: %s", td)
	out.Debug(ctx, "URL: %s", u.String())
	archive := filepath.Join(td, path.Base(u.Path))
	if err := s.tryDownload(ctx, archive, u.String()); err != nil {
		return err
	}
	binDst := exe + "_new"
	_ = os.Remove(binDst)
	if err := s.extract(ctx, archive, binDst); err != nil {
		return err
	}

	// for tests
	if !updateMoveAfterQuit {
		return nil
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
	// check if this is a test binary
	if filepath.Base(fn) == "action.test" {
		return nil
	}
	// check if we want to force updateability
	if uf := os.Getenv("GOPASS_FORCE_UPDATE"); uf != "" {
		out.Debug(ctx, "updateable due to force flag")
		return nil
	}
	// check if file is in GOPATH
	if gp := os.Getenv("GOPATH"); strings.HasPrefix(fn, gp) {
		return fmt.Errorf("use go get -u to update binary in GOPATH")
	}
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
