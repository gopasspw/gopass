// +build !windows

package sub

import "context"

func (s *Store) gitFixConfigOSDep(ctx context.Context) error {
	// nothing to do
	return nil
}
