package gpgconf

import (
	"context"
)

// Binary returns the GPG binary location.
func Binary(ctx context.Context, bin string) (string, error) {
	return detectBinary(ctx, bin)
}
