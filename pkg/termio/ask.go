package termio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
)

var (
	// Stderr is exported for tests.
	Stderr io.Writer = os.Stderr
	// Stdin is exported for tests.
	Stdin io.Reader = os.Stdin
	// ErrAborted is returned if the user aborts an action.
	ErrAborted = fmt.Errorf("user aborted")
	// ErrInvalidInput is returned if the user enters invalid input.
	ErrInvalidInput = fmt.Errorf("no valid user input")
)

const (
	maxTries = 42
)

// AskForString asks for a string once, using the default if the
// answer is empty. Errors are only returned on I/O errors.
func AskForString(ctx context.Context, text, def string) (string, error) {
	if ctxutil.IsAlwaysYes(ctx) || !ctxutil.IsInteractive(ctx) {
		return def, nil
	}

	// check for context cancelation
	select {
	case <-ctx.Done():
		return def, ErrAborted
	default:
	}

	fmt.Fprintf(Stderr, "%s [%s]: ", text, def)

	input, err := NewReader(ctx, Stdin).ReadLine()
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		input = def
	}

	return input, nil
}

// AskForBool ask for a bool (yes or no) exactly once.
// The empty answer uses the specified default, any other answer
// is an error.
func AskForBool(ctx context.Context, text string, def bool) (bool, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return def, nil
	}

	choices := "y/N/q"
	if def {
		choices = "Y/n/q"
	}

	str, err := AskForString(ctx, text, choices)
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	switch str {
	case "Y/n/q":
		return true, nil
	case "y/N/q":
		return false, nil
	}

	str = strings.ToLower(string(str[0]))
	switch str {
	case "y":
		return true, nil
	case "n":
		return false, nil
	case "q":
		return false, ErrAborted
	default:
		return false, fmt.Errorf("unknown answer '%s': %w", str, ErrInvalidInput)
	}
}

// AskForInt asks for an valid interger once. If the input
// can not be converted to an int it returns an error.
func AskForInt(ctx context.Context, text string, def int) (int, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return def, nil
	}

	str, err := AskForString(ctx, text+" (q to abort)", strconv.Itoa(def))
	if err != nil {
		return 0, err
	}

	if str == "q" {
		return 0, ErrAborted
	}

	intVal, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("failed to convert to number: %w", err)
	}

	return intVal, nil
}

// AskForConfirmation asks a yes/no question until the user
// replies yes or no.
func AskForConfirmation(ctx context.Context, text string) bool {
	if ctxutil.IsAlwaysYes(ctx) {
		return true
	}

	for range maxTries {
		choice, err := AskForBool(ctx, text, false)
		if err == nil {
			return choice
		}

		if errors.Is(err, ErrAborted) {
			return false
		}
	}

	return false
}

// AskForKeyImport asks for permissions to import the named key.
func AskForKeyImport(ctx context.Context, key string, names []string) bool {
	if ctxutil.IsAlwaysYes(ctx) {
		return true
	}

	if !ctxutil.IsInteractive(ctx) {
		return false
	}

	ok, err := AskForBool(ctx, fmt.Sprintf("Do you want to import the public key %q (Names: %+v) into your keyring?", key, names), false)
	if err != nil {
		return false
	}

	return ok
}

// AskForPassword prompts for a password, optionally prompting twice until both match.
func AskForPassword(ctx context.Context, name string, repeat bool) (string, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return "", nil
	}

	askFn := GetPassPromptFunc(ctx)

	for range maxTries {
		// check for context cancellation
		select {
		case <-ctx.Done():
			return "", ErrAborted
		default:
		}

		pass, err := askFn(ctx, fmt.Sprintf("Enter %s", name))
		if !repeat {
			return pass, err
		}

		if err != nil {
			return "", err
		}

		passAgain, err := askFn(ctx, fmt.Sprintf("Retype %s", name))
		if err != nil {
			return "", err
		}

		if pass == passAgain {
			return pass, nil
		}

		out.Errorf(ctx, "Error: the entered password do not match")
	}

	return "", ErrInvalidInput
}
