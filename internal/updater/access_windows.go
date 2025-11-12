//go:build windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

func canWrite(path string) error {
	return nil
}

// Windows won't allow us to remove the binary that's currently being executed.
// So rename the binary and then the updater should be able to write it's
// update to the correct location.
//
// See https://stackoverflow.com/a/459860
func removeOldBinary(dir, dest string) error {
	bakFile := filepath.Join(dir, filepath.Base(dest)+".bak")
	// check if the bakup file already exists
	if _, err := os.Stat(bakFile); err == nil {
		// ... then remove it
		_ = os.Remove(bakFile)
	}
	// we can't remove the currently running binary, but should be able to
	// rename it.
	if err := os.Rename(dest, bakFile); err != nil {
		return fmt.Errorf("unable to rename %s to %s: %w", dest, bakFile, err)
	}

	return nil
}
