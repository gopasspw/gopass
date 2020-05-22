package fs

import (
	"os"
	"syscall"
)

func notEmptyErr(err error) bool {
	return err.(*os.PathError).Err == syscall.ERROR_DIR_NOT_EMPTY
}
