package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

// IsUpdateable returns an error if this binary is not updateable.
//
//nolint:goerr113
func IsUpdateable(ctx context.Context) error {
	fn, err := executable(ctx)
	if err != nil {
		return err
	}

	debug.Log("File: %s", fn)
	// check if this is a test binary
	if strings.HasSuffix(filepath.Base(fn), ".test") {
		return nil
	}

	// check if we want to force updateability
	if uf := os.Getenv("GOPASS_FORCE_UPDATE"); uf != "" {
		debug.Log("updateable due to force flag")

		return nil
	}

	// check if file is in GOPATH
	if gp := os.Getenv("GOPATH"); strings.HasPrefix(fn, gp) {
		return fmt.Errorf("use go get -u to update binary in GOPATH")
	}

	// check file
	fi, err := os.Stat(fn)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if !fi.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	}

	if err := canWrite(fn); err != nil {
		return fmt.Errorf("can not write %q: %w", fn, err)
	}

	// no need to check the directory since we'll be writing to the destination file directly
	return nil
}

//nolint:wrapcheck
var executable = func(ctx context.Context) (string, error) {
	path, err := os.Executable()
	if err != nil {
		return path, err
	}
	path, err = filepath.EvalSymlinks(path)

	return path, err
}
