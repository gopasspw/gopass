//go:build !darwin

package clipboard

import (
	"context"
	"fmt"

	"github.com/gopasspw/clipboard"
)

func copyToClipboard(ctx context.Context, content []byte) error {
	// We should be using clipboard.WritePassword here, but many
	// Linux distros currently do not ship with the required dependencies.
	// See https://github.com/gopasspw/gopass/pull/3234
	if err := clipboard.WriteAll(ctx, content); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	return nil
}
