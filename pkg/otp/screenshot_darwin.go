//go:build !cgo

package otp

import (
	"context"
	"fmt"
)

// ParseScreen will attempt to parse all available screen and will look for otpauth QR codes. It returns the first one
// it has found.
func ParseScreen(ctx context.Context) (string, error) {
	return "", fmt.Errorf("Not supported on OSX without compiling with CGO")
}
