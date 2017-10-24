// +build windows

package sub

import (
	"context"

	"github.com/pkg/errors"
)

func (s *Store) gitFixConfigOSDep(ctx context.Context) error {
	if err := s.gitCmd(ctx, "gitFixConfigOSDep", "config", "--local", "gpg.program", "TODO"); err != nil {
		return errors.Wrapf(err, "failed to set git config gpg.program")
	}
	return nil
}
