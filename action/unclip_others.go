// +build !linux

package action

import "context"

func (s *Action) clearClipboardHistory(ctx context.Context) error {
	return nil
}

func (s *Action) unclipNotify(ctx context.Context, msg string) error {
	return nil
}
