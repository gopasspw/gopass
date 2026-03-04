//go:build !windows

package fs

import (
	"errors"
	"os"
	"syscall"
)

func notEmptyErr(err error) bool {
	var perr *os.PathError
	if errors.As(err, &perr) {
		return errors.Is(perr.Err, syscall.ENOTEMPTY)
	}

	return false
}
