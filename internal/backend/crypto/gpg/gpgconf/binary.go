package gpgconf

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Binary returns the GPG binary location.
func Binary(ctx context.Context, bin string) (string, error) {
	if sv := os.Getenv("GOPASS_GPG_BINARY"); sv != "" {
		debug.Log("Using GOPASS_GPG_BINARY: %s", sv)

		return sv, nil
	}

	return detectBinary(ctx, bin)
}
