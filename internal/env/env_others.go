//go:build !darwin

package env

import "context"

// Check does nothing on these OSes, yet.
func Check(ctx context.Context) (string, error) {
	return "", nil
}
