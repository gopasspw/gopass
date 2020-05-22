// +build windows

package termio

import (
	"context"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/pkg/errors"
)

// promptPass will prompt user's for a password by terminal.
func promptPass(ctx context.Context, prompt string) (string, error) {
	if !ctxutil.IsTerminal(ctx) {
		return AskForString(ctx, prompt, "")
	}

	return "", errors.New("not a terminal")
}
