//go:build windows
// +build windows

package termio

import (
	"context"
	"fmt"
	"os"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"golang.org/x/crypto/ssh/terminal"
)

// promptPass will prompt user's for a password by terminal.
func promptPass(ctx context.Context, prompt string) (string, error) {
	if !ctxutil.IsTerminal(ctx) {
		return AskForString(ctx, prompt, "")
	}

	fmt.Fprintf(Stdout, "%s: ", prompt)
	passBytes, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(Stdout, "")
	return string(passBytes), err
}
