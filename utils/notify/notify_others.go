// +build !linux,!windows,!darwin

package notify

import (
	"runtime"

	"github.com/pkg/errors"
)

// Notify is not yet implemented on this platform
func Notify(subj, msg string) error {
	return errors.Errorf("GOOS %s not yet supported", runtime.GOOS)
}
