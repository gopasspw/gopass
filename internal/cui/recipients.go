package cui

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
)

var (
	// Stdin is exported for tests.
	Stdin io.Reader = os.Stdin
	// Stdout is exported for tests.
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests.
	Stderr io.Writer = os.Stderr
)

const (
	maxTries = 42
)

// AskForPrivateKey prompts the user to select from a list of private keys.
func AskForPrivateKey(ctx context.Context, crypto backend.Crypto, prompt string) (string, error) {
	if !ctxutil.IsInteractive(ctx) {
		return "", fmt.Errorf("can not select private key without terminal")
	}

	if crypto == nil {
		return "", fmt.Errorf("can not select private key without valid crypto backend")
	}

	kl, err := crypto.ListIdentities(ctx)
	if err != nil {
		return "", err
	}

	if len(kl) < 1 {
		return "", fmt.Errorf("no useable private keys found. make sure you have valid private keys with sufficient trust")
	}

	// shortcut: If there is only one key, use it
	if len(kl) == 1 {
		return kl[0], nil
	}

	fmtStr := "[%" + strconv.Itoa((len(kl)/10)+1) + "d] %s - %s\n"
	for i := 0; i < maxTries; i++ {
		if !ctxutil.IsTerminal(ctx) {
			return kl[0], nil
		}
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("user aborted")
		default:
		}

		fmt.Fprintln(Stdout, prompt)
		for i, k := range kl {
			fmt.Fprintf(Stdout, fmtStr, i, crypto.Name(), crypto.FormatKey(ctx, k, ""))
		}

		iv, err := termio.AskForInt(ctx, fmt.Sprintf("Please enter the number of a key (0-%d, [q]uit)", len(kl)-1), 0)
		if err != nil {
			if err.Error() == "user aborted" {
				return "", err
			}

			continue
		}

		if iv >= 0 && iv < len(kl) {
			return kl[iv], nil
		}
	}

	return "", fmt.Errorf("no valid user input")
}

// AskForGitConfigUser will iterate over GPG private key identities and prompt
// the user for selecting one identity whose name and email address will be used as
// git config user.name and git config user.email, respectively.
// On error or no selection, name and email will be empty.
//
// If s.isTerm is false (i.e., the user cannot be prompted), however,
// the first identity's name/email pair found is returned.
func AskForGitConfigUser(ctx context.Context, crypto backend.Crypto) (string, string, error) {
	var useCurrent bool

	if crypto == nil {
		return "", "", fmt.Errorf("crypto not available")
	}

	if crypto.Name() == "age" {
		debug.Log("skipping git config user prompt for non-gpg backend %s", crypto.Name())

		return "", "", nil
	}

	keyList, err := crypto.ListIdentities(ctx)
	if err != nil {
		return "", "", err
	}

	if len(keyList) < 1 {
		return "", "", fmt.Errorf("no usable private keys found")
	}

	for _, key := range keyList {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", "", fmt.Errorf("user aborted")
		default:
		}

		name := crypto.FormatKey(ctx, key, "{{ .Identity.Name }}")
		email := crypto.FormatKey(ctx, key, "{{ .Identity.Email }}")

		if name == "" && email == "" {
			continue
		}

		useCurrent, err = termio.AskForBool(
			ctx,
			fmt.Sprintf("Use %s (%s) for password store git config?", name, email),
			true,
		)

		if err != nil {
			return "", "", err
		}

		if useCurrent {
			return name, email, nil
		}
	}

	return "", "", nil
}

type mountPointer interface {
	MountPoints() []string
}

func sorted(s []string) []string {
	sort.Strings(s)

	return s
}

// AskForStore shows a store / mount point selection.
func AskForStore(ctx context.Context, s mountPointer) string {
	if !ctxutil.IsInteractive(ctx) {
		return ""
	}

	stores := []string{"<root>"}
	stores = append(stores, sorted(s.MountPoints())...)
	if len(stores) < 2 {
		return ""
	}

	act, sel := GetSelection(ctx, "Please select the store you would like to use", stores)
	switch act {
	case "default":
		fallthrough
	case "show":
		store := stores[sel]
		if store == "<root>" {
			store = ""
		}

		return store
	default:
		return "" // root store
	}
}
