// +build windows

package action

import (
	"context"

	"github.com/pkg/errors"
)

func (s *Action) updateGopass(ctx context.Context, version, url string) error {
	return errors.Errorf("Windows is not yet supported")
}

func (s *Action) isUpdateable(ctx context.Context) error {
	return errors.Errorf("Windows is not yet supported")
}
