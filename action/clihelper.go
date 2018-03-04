package action

import (
	"context"
	"fmt"
	"sort"

	"github.com/justwatchcom/gopass/backend/crypto/gpg"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/cui"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termio"
	"github.com/pkg/errors"
)

const (
	maxTries = 42
)

// ConfirmRecipients asks the user to confirm a given set of recipients
func (s *Action) ConfirmRecipients(ctx context.Context, name string, recipients []string) ([]string, error) {
	if ctxutil.IsNoConfirm(ctx) || !ctxutil.IsInteractive(ctx) {
		return recipients, nil
	}

	crypto := s.Store.Crypto(ctx, name)
	sort.Strings(recipients)

	fmt.Fprintf(stdout, "gopass: Encrypting %s for these recipients:\n", name)
	for _, r := range recipients {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return nil, errors.New("user aborted")
		default:
		}

		kl, err := crypto.FindPublicKeys(ctx, r)
		if err != nil {
			out.Red(ctx, "Failed to read public key for '%s': %s", name, err)
			continue
		}
		if len(kl) < 1 {
			fmt.Fprintln(stdout, "key not found", r)
			continue
		}
		fmt.Fprintf(stdout, " - %s\n", crypto.FormatKey(ctx, kl[0]))
	}
	fmt.Fprintln(stdout, "")

	yes, err := termio.AskForBool(ctx, "Do you want to continue?", true)
	if err != nil {
		return recipients, errors.Wrapf(err, "failed to read user input")
	}
	if yes {
		return recipients, nil
	}

	return recipients, errors.New("user aborted")
}

// askforPrivateKey promts the user to select from a list of private keys
func (s *Action) askForPrivateKey(ctx context.Context, name, prompt string) (string, error) {
	if !ctxutil.IsInteractive(ctx) {
		return "", errors.New("no interaction without terminal")
	}

	crypto := s.Store.Crypto(ctx, name)
	kl, err := crypto.ListPrivateKeyIDs(gpg.WithAlwaysTrust(ctx, false))
	if err != nil {
		return "", err
	}
	if len(kl) < 1 {
		return "", errors.New("No useable private keys found")
	}

	for i := 0; i < maxTries; i++ {
		if !ctxutil.IsTerminal(ctx) {
			return kl[0], nil
		}
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", errors.New("user aborted")
		default:
		}

		fmt.Fprintln(stdout, prompt)
		for i, k := range kl {
			fmt.Fprintf(stdout, "[%d] %s\n", i, crypto.FormatKey(ctx, k))
		}
		iv, err := termio.AskForInt(ctx, fmt.Sprintf("Please enter the number of a key (0-%d, [q]uit)", len(kl)-1), 0)
		if err != nil {
			continue
		}
		if iv >= 0 && iv < len(kl) {
			return kl[iv], nil
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
func (s *Action) askForGitConfigUser(ctx context.Context, name string) (string, string, error) {
	var useCurrent bool

	crypto := s.Store.Crypto(ctx, name)
	keyList, err := crypto.ListPrivateKeyIDs(ctx)
	if err != nil {
		return "", "", err
	}
	if len(keyList) < 1 {
		return "", "", errors.New("No usable private keys found")
	}

	for _, key := range keyList {
		// check for context cancelation
		select {
		case <-ctx.Done():
			return "", "", errors.New("user aborted")
		default:
		}

		name := crypto.NameFromKey(ctx, key)
		email := crypto.EmailFromKey(ctx, key)

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

func (s *Action) askForStore(ctx context.Context) string {
	if !ctxutil.IsInteractive(ctx) {
		return ""
	}

	mps := s.Store.MountPoints()
	if len(mps) < 1 {
		return ""
	}

	stores := []string{"<root>"}
	stores = append(stores, mps...)
	act, sel := cui.GetSelection(ctx, "Please select the store you would like to use", "<↑/↓> to change the selection, <→> to select, <ESC> to quit", stores)
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
