package termio

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/pkg/errors"
)

var (
	// Stdout is exported for tests
	Stdout io.Writer = os.Stdout
	// Stdin is exported for tests
	Stdin io.Reader = os.Stdin
)

const (
	maxTries = 42
)

// AskForString asks for a string once, using the default if the
// anser is empty. Errors are only returned on I/O errors
func AskForString(ctx context.Context, text, def string) (string, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return def, nil
	}

	// check for context cancelation
	select {
	case <-ctx.Done():
		return def, errors.New("user aborted")
	default:
	}

	fmt.Fprintf(Stdout, "%s [%s]: ", text, def)
	input, err := NewReader(Stdin).ReadLine()
	if err != nil {
		return "", errors.Wrapf(err, "failed to read user input")
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
		return false, errors.Wrapf(err, "failed to read user input")
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
		return false, errors.Errorf("user aborted")
	default:
		return false, errors.Errorf("Unknown answer: %s", str)
	}
}

// AskForInt asks for an valid interger once. If the input
// can not be converted to an int it returns an error
func AskForInt(ctx context.Context, text string, def int) (int, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return def, nil
	}

	str, err := AskForString(ctx, text, strconv.Itoa(def))
	if err != nil {
		return 0, err
	}
	if str == "q" {
		return 0, errors.Errorf("user aborted")
	}
	intVal, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert to number")
	}
	return intVal, nil
}

// AskForConfirmation asks a yes/no question until the user
// replies yes or no
func AskForConfirmation(ctx context.Context, text string) bool {
	if ctxutil.IsAlwaysYes(ctx) {
		return true
	}

	for i := 0; i < maxTries; i++ {
		if choice, err := AskForBool(ctx, text, false); err == nil {
			return choice
		}
	}
	return false
}

// AskForKeyImport asks for permissions to import the named key
func AskForKeyImport(ctx context.Context, key string, names []string) bool {
	if ctxutil.IsAlwaysYes(ctx) {
		return true
	}
	if !ctxutil.IsInteractive(ctx) {
		return false
	}

	ok, err := AskForBool(ctx, fmt.Sprintf("Do you want to import the public key '%s' (Names: %+v) into your keyring?", key, names), false)
	if err != nil {
		return false
	}

	return ok
}

// AskForPassword prompts for a password twice until both match
func AskForPassword(ctx context.Context, name string) (string, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return "", nil
	}

	askFn := GetPassPromptFunc(ctx)
	for i := 0; i < maxTries; i++ {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", errors.New("user aborted")
		default:
		}

		pass, err := askFn(ctx, fmt.Sprintf("Enter password for %s", name))
		if err != nil {
			return "", err
		}

		passAgain, err := askFn(ctx, fmt.Sprintf("Retype password for %s", name))
		if err != nil {
			return "", err
		}

		if pass == passAgain || pass == "" {
			return pass, nil
		}

		out.Red(ctx, "Error: the entered password do not match")
	}
	return "", errors.New("no valid user input")
}
