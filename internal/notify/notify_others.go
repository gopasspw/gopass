//go:build !linux && !windows && !darwin
// +build !linux,!windows,!darwin

package notify

import (
	"context"
	"fmt"
	"runtime"
)

// Notify is not yet implemented on this platform
func Notify(ctx context.Context, subj, msg string) error {
	return fmt.Errorf("GOOS %s not yet supported", runtime.GOOS)
}
