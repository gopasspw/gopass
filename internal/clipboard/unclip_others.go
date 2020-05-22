// +build !linux

package clipboard

import "context"

func clearClipboardHistory(ctx context.Context) error {
	return nil
}
