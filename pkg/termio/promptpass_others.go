// +build !windows

package termio

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// promptPass will prompt user's for a password by terminal.
func promptPass(ctx context.Context, prompt string) (string, error) {
	if !ctxutil.IsTerminal(ctx) {
		return AskForString(ctx, prompt, "")
	}

	// Make a copy of STDIN's state to restore afterward
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.GetState(fd)
	if err != nil {
		return "", errors.Errorf("Could not get state of terminal: %s", err)
	}
	defer func() {
		if err := terminal.Restore(fd, oldState); err != nil {
			out.Red(ctx, "Failed to restore terminal: %s", err)
		}
	}()

	// Restore STDIN in the event of a signal interruption
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		<-sigch
		if err := terminal.Restore(fd, oldState); err != nil {
			out.Red(ctx, "Failed to restore terminal: %s", err)
		}
		os.Exit(1)
	}()

	fmt.Fprintf(Stdout, "%s: ", prompt)
	passBytes, err := terminal.ReadPassword(fd)
	fmt.Fprintln(Stdout, "")
	return string(passBytes), err
}
