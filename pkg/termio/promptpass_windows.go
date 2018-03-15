// +build windows

package termio

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
)

// promptPass will prompt user's for a password by terminal.
func promptPass(ctx context.Context, prompt string) (string, error) {
	if !ctxutil.IsTerminal(ctx) {
		return "", nil
	}

	return AskForString(ctx, prompt, "")
}
