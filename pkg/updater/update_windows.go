// +build windows

package updater

import (
	"context"

	"github.com/pkg/errors"
)

func updateGopass(ctx context.Context, version, url string) error {
	return errors.Errorf("Windows is not yet supported")
}

func IsUpdateable(ctx context.Context) error {
	return errors.Errorf("Windows is not yet supported")
}
