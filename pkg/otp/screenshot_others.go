//go:build !(arm || arm64 || amd64 || 386) || !(linux || windows || freebsd || netbsd) || (darwin && !cgo)
// +build !arm,!arm64,!amd64,!386 !linux,!windows,!freebsd,!netbsd darwin,!cgo

package otp

import (
	"context"
	"fmt"
)

// ParseScreen will attempt to parse all available screen and will look for otpauth QR codes. It returns the first one
// it has found.
func ParseScreen(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not supported on your platform")
}
