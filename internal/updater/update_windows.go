// +build windows

package updater

import (
	"context"
	"fmt"
)

func IsUpdateable(ctx context.Context) error {
	return fmt.Errorf("Windows is not yet supported")
}

var executable = func(ctx context.Context) (string, error) {
	return "", nil
}
