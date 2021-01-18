// +build windows

package updater

import (
	"context"

	"github.com/pkg/errors"
)

func IsUpdateable(ctx context.Context) error {
	return errors.Errorf("Windows is not yet supported")
}

var executable = func(ctx context.Context) (string, error) {
	return "", nil
}
