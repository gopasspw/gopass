//go:build !windows
// +build !windows

package fs

import (
	"os"
	"syscall"
)

func notEmptyErr(err error) bool {
	e, ok := err.(*os.PathError)
	if !ok {
		return false
	}
	return e.Err == syscall.ENOTEMPTY
}
