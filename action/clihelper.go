package action

import (
	"context"
	"fmt"
	"sort"

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
		iv, err := termio.AskForInt(ctx, fmt.Sprintf("Please enter the number of a key (0-%d)", len(kl)-1), 0)
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

			useCurrent, err = termio.AskForBool(
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
