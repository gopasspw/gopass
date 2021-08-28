//go:build !windows
// +build !windows

package fs

import (
	"os"
	"syscall"
)

func notEmptyErr(err error) bool {
	return err.(*os.PathError).Err == syscall.ENOTEMPTY
}
