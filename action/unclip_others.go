// +build !linux

package action

import "context"

func (s *Action) clearClipboardHistory(ctx context.Context) error {
	return nil
}
