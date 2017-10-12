// +build windows

package action

import (
	"context"

	"github.com/justwatchcom/gopass/utils/ctxutil"
)

// promptPass will prompt user's for a password by terminal.
func (s *Action) promptPass(ctx context.Context, prompt string) (string, error) {
	if !ctxutil.IsTerminal(ctx) {
		return "", nil
	}

	return s.askForString(ctx, prompt, "")
}
