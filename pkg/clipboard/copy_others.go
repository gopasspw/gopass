//go:build !darwin
// +build !darwin

package clipboard

import (
	"context"

	"github.com/atotto/clipboard"
)

func copyToClipboard(ctx context.Context, content []byte) error {
	return clipboard.WriteAll(string(content))
}
