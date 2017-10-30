// +build !linux

package action

import "context"

func (s *Action) clearClipboardHistory(ctx context.Context) error {
	return nil
}

func (s *Action) desktopNotify(ctx context.Context, subj, msg string) error {
	return nil
}
