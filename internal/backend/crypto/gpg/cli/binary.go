package cli

import (
	"context"
)

// Binary returns the GPG binary location
func (g *GPG) Binary() string {
	if g == nil {
		return ""
	}
	return g.binary
}

// Binary returns the GPG binary location
func Binary(ctx context.Context, bin string) (string, error) {
	return detectBinary(bin)
}
