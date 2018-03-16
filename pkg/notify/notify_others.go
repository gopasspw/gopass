// +build !linux,!windows,!darwin

package notify

import (
	"context"
	"runtime"

	"github.com/pkg/errors"
)

// Notify is not yet implemented on this platform
func Notify(ctx context.Context, subj, msg string) error {
	return errors.Errorf("GOOS %s not yet supported", runtime.GOOS)
}
