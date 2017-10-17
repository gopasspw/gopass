// +build !windows

package cli

import "context"

func (g *Git) fixConfigOSDep(ctx context.Context) error {
	// nothing to do
	return nil
}
