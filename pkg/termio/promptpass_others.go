//go:build !windows
// +build !windows

package termio

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/term"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
)

// promptPass will prompt user's for a password by terminal.
func promptPass(ctx context.Context, prompt string) (string, error) {
	if !ctxutil.IsTerminal(ctx) {
		return AskForString(ctx, prompt, "")
	}

	// Make a copy of STDIN's state to restore afterward
	fd := int(os.Stdin.Fd())
	oldState, err := term.GetState(fd)
	if err != nil {
		return "", fmt.Errorf("could not get state of terminal: %w", err)
	}
	defer func() {
		if err := term.Restore(fd, oldState); err != nil {
			out.Errorf(ctx, "Failed to restore terminal: %s", err)
		}
	}()

	// Restore STDIN in the event of a signal interruption
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		<-sigch
		if err := term.Restore(fd, oldState); err != nil {
			out.Errorf(ctx, "Failed to restore terminal: %s", err)
		}
		os.Exit(1)
	}()

	fmt.Fprintf(Stderr, "%s: ", prompt)
	passBytes, err := term.ReadPassword(fd)
	fmt.Fprintln(Stderr, "")
	return string(passBytes), err
}
