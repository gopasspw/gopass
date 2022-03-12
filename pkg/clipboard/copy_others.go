//go:build !darwin
// +build !darwin

package clipboard

import (
	"context"
	"fmt"

	"github.com/atotto/clipboard"
)

func copyToClipboard(ctx context.Context, content []byte) error {
	if err := clipboard.WriteAll(string(content)); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	return nil
}
