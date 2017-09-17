// +build !windows

package action

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// promptPass will prompt user's for a password by terminal.
func (s *Action) promptPass(ctx context.Context, prompt string) (pass string, err error) {
	if !ctxutil.IsTerminal(ctx) {
		return
	}

	// Make a copy of STDIN's state to restore afterward
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.GetState(fd)
	if err != nil {
		return "", errors.Errorf("Could not get state of terminal: %s", err)
	}
	defer func() {
		if err := terminal.Restore(fd, oldState); err != nil {
			fmt.Println(color.RedString("Failed to restore terminal: %s", err))
		}
	}()

	// Restore STDIN in the event of a signal interruption
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		for range sigch {
			if err := terminal.Restore(fd, oldState); err != nil {
				fmt.Println(color.RedString("Failed to restore terminal: %s", err))
			}
			os.Exit(1)
		}
	}()

	fmt.Printf("%s: ", prompt)
	passBytes, err := terminal.ReadPassword(fd)
	fmt.Println("")
	return string(passBytes), err
}
