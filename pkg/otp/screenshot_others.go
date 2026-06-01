//go:build !((arm || arm64 || amd64 || 386) && (linux || windows || (cgo && darwin) || freebsd || netbsd)) || noscreenshot

package otp

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// ParseScreen will attempt to parse all available screen and will look for otpauth QR codes. It returns the first one
// it has found.
func ParseScreen(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not supported on your platform")
}

// SnipFlags returns an empty slice because the screen-capture feature is not available in this build.
func SnipFlags() []cli.Flag {
	return nil
}
