package action

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termwiz"
	"github.com/pkg/errors"
)

const (
	maxTries = 42
)

// ConfirmRecipients asks the user to confirm a given set of recipients
func (s *Action) ConfirmRecipients(ctx context.Context, name string, recipients []string) ([]string, error) {
	if ctxutil.IsNoConfirm(ctx) || !ctxutil.IsInteractive(ctx) || ctxutil.IsAlwaysYes(ctx) {
		return recipients, nil
	}

	sort.Strings(recipients)

	fmt.Printf("gopass: Encrypting %s for these recipients:\n", name)
	for _, r := range recipients {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return nil, errors.New("user aborted")
		default:
		}

		kl, err := s.gpg.FindPublicKeys(ctx, r)
		if err != nil {
			out.Red(ctx, "Failed to read public key for '%s': %s", name, err)
			continue
		}
		if len(kl) < 1 {
			fmt.Println("key not found", r)
			continue
		}
		fmt.Printf(" - %s\n", kl[0].OneLine())
	}
	fmt.Println("")

	yes, err := s.askForBool(ctx, "Do you want to continue?", true)
	if err != nil {
		return recipients, errors.Wrapf(err, "failed to read user input")
	}

	if yes {
		return recipients, nil
	}

	return recipients, errors.New("user aborted")
}

// AskForConfirmation asks a yes/no question until the user
// replies yes or no
func (s *Action) AskForConfirmation(ctx context.Context, text string) bool {
	if ctxutil.IsAlwaysYes(ctx) {
		return true
	}

	for i := 0; i < maxTries; i++ {
		if choice, err := s.askForBool(ctx, text, false); err == nil {
			return choice
		}
	}
	return false
}

// askForBool ask for a bool (yes or no) exactly once.
// The empty answer uses the specified default, any other answer
// is an error.
func (s *Action) askForBool(ctx context.Context, text string, def bool) (bool, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return def, nil
	}

	choices := "y/N"
	if def {
		choices = "Y/n"
	}

	str, err := s.askForString(ctx, text, choices)
	if err != nil {
		return false, errors.Wrapf(err, "failed to read user input")
	}
	switch str {
	case "Y/n":
		return true, nil
	case "y/N":
		return false, nil
	}

	str = strings.ToLower(string(str[0]))
	switch str {
	case "y":
		return true, nil
	case "n":
		return false, nil
	default:
		return false, errors.Errorf("Unknown answer: %s", str)
	}
}

// askForString asks for a string once, using the default if the
// anser is empty. Errors are only returned on I/O errors
func (s *Action) askForString(ctx context.Context, text, def string) (string, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return def, nil
	}

	// check for context cancelation
	select {
	case <-ctx.Done():
		return def, errors.New("user aborted")
	default:
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [%s]: ", text, def)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.Wrapf(err, "failed to read user input")
	}
	input = strings.TrimSpace(input)
	if input == "" {
		input = def
	}
	return input, nil
}

// askForInt asks for an valid interger once. If the input
// can not be converted to an int it returns an error
func (s *Action) askForInt(ctx context.Context, text string, def int) (int, error) {
	if ctxutil.IsAlwaysYes(ctx) {
		return def, nil
	}

	str, err := s.askForString(ctx, text, strconv.Itoa(def))
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert to number")
	}
	return intVal, nil
}

// askForPassword prompts for a password twice until both match
func (s *Action) askForPassword(ctx context.Context, name string, askFn func(context.Context, string) (string, error)) (string, error) {
	if !ctxutil.IsInteractive(ctx) {
		return "", errors.New("impossible without terminal")
	}
	if ctxutil.IsAlwaysYes(ctx) {
		return "", nil
	}

	if askFn == nil {
		askFn = s.promptPass
	}
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

		if pass == passAgain {
			return strings.TrimSpace(pass), nil
		}

		fmt.Println("Error: the entered password do not match")
	}
	return "", errors.New("no valid user input")
}

// AskForKeyImport asks for permissions to import the named key
func (s *Action) AskForKeyImport(ctx context.Context, key string, names []string) bool {
	if ctxutil.IsAlwaysYes(ctx) {
		return true
	}
	if !ctxutil.IsInteractive(ctx) {
		return false
	}

	ok, err := s.askForBool(ctx, fmt.Sprintf("Do you want to import the public key '%s' (Names: %+v) into your keyring?", key, names), false)
	if err != nil {
		return false
	}

	return ok
}

// askforPrivateKey promts the user to select from a list of private keys
func (s *Action) askForPrivateKey(ctx context.Context, prompt string) (string, error) {
	if !ctxutil.IsInteractive(ctx) {
		return "", errors.New("no interaction without terminal")
	}
	kl, err := s.gpg.ListPrivateKeys(ctx)
	if err != nil {
		return "", err
	}
	kl = kl.UseableKeys()
	if len(kl) < 1 {
		return "", errors.New("No useable private keys found")
	}
	for i := 0; i < maxTries; i++ {
		if ctxutil.IsAlwaysYes(ctx) {
			return kl[0].Fingerprint, nil
		}
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", errors.New("user aborted")
		default:
		}

		fmt.Println(prompt)
		for i, k := range kl {
			fmt.Printf("[%d] %s\n", i, k.OneLine())
		}
		iv, err := s.askForInt(ctx, fmt.Sprintf("Please enter the number of a key (0-%d)", len(kl)-1), 0)
		if err != nil {
			continue
		}
		if iv >= 0 && iv < len(kl) {
			return kl[iv].Fingerprint, nil
		}
	}
	return "", errors.New("no valid user input")
}

// askForGitConfigUser will iterate over GPG private key identities and prompt
// the user for selecting one identity whose name and email address will be used as
// git config user.name and git config user.email, respectively.
// On error or no selection, name and email will be empty.
// If s.isTerm is false (i.e., the user cannot be prompted), however,
// the first identity's name/email pair found is returned.
func (s *Action) askForGitConfigUser(ctx context.Context) (string, string, error) {
	var useCurrent bool

	keyList, err := s.gpg.ListPrivateKeys(ctx)
	if err != nil {
		return "", "", err
	}
	keyList = keyList.UseableKeys()
	if len(keyList) < 1 {
		return "", "", errors.New("No usable private keys found")
	}

	for _, key := range keyList {
		for _, identity := range key.Identities {
			if !ctxutil.IsTerminal(ctx) || ctxutil.IsAlwaysYes(ctx) {
				return identity.Name, identity.Email, nil
			}
			// check for context cancelation
			select {
			case <-ctx.Done():
				return "", "", errors.New("user aborted")
			default:
			}

			useCurrent, err = s.askForBool(
				ctx,
				fmt.Sprintf("Use %s (%s) for password store git config?", identity.Name, identity.Email), true)
			if err != nil {
				return "", "", err
			}
			if useCurrent {
				return identity.Name, identity.Email, nil
			}
		}
	}

	return "", "", nil
}

func (s *Action) askForStore(ctx context.Context) string {
	mps := s.Store.MountPoints()
	if len(mps) < 1 {
		return ""
	}

	stores := []string{"<root>"}
	stores = append(stores, mps...)
	act, sel := termwiz.GetSelection(ctx, "Store for secret", "<↑/↓> to change the selection, <→> to select, <ESC> to quit", stores)
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
