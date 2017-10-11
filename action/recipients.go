package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/termwiz"
	"github.com/urfave/cli"
)

var (
	removalWarning = `


@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@ WARNING: REMOVING A USER WILL NOT REVOKE ACCESS FROM OLD REVISONS!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
THE USER %s WILL STILL BE ABLE TO ACCESS ANY OLD COPY OF THE STORE AND
ANY OLD REVISION HE HAD ACCESS TO.

ANY CREDENTIALS THIS USER HAD ACCESS TO NEED TO BE CONSIDERED COMPROMISED
AND SHOULD BE REVOKED.

This feature is only meant from revoking access to any added or changed
credentials.

`
)

// RecipientsPrint prints all recipients per store
func (s *Action) RecipientsPrint(ctx context.Context, c *cli.Context) error {
	if err := s.Store.ImportMissingPublicKeys(ctx); err != nil {
		out.Red(ctx, "Failed to import missing public keys: %s", err)
	}

	if err := s.Store.SaveRecipients(ctx); err != nil {
		out.Red(ctx, "Failed to export missing public keys: %s", err)
	}

	tree, err := s.Store.RecipientsTree(ctx, true)
	if err != nil {
		return s.exitError(ctx, ExitList, err, "failed to list recipients: %s", err)
	}

	fmt.Println(tree.Format(0))
	return nil
}

// RecipientsComplete will print a list of recipients for bash
// completion
func (s *Action) RecipientsComplete(ctx context.Context, c *cli.Context) {
	tree, err := s.Store.RecipientsTree(ctx, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range tree.List(0) {
		fmt.Println(v)
	}
}

// RecipientsAdd adds new recipients
func (s *Action) RecipientsAdd(ctx context.Context, c *cli.Context) error {
	store := c.String("store")
	added := 0

	recipients := []string(c.Args())
	if len(recipients) < 1 {
		choices := []string{}
		kl, _ := s.gpg.FindPublicKeys(ctx)
		for _, key := range kl.UseableKeys() {
			choices = append(choices, key.OneLine())
		}
		if len(choices) > 0 {
			act, sel := termwiz.GetSelection(ctx, "Add Recipient -", "<↑/↓> to change the selection, <→> to add this recipient, <ESC> to quit", choices)
			switch act {
			case "show":
				recipients = []string{choices[sel]}
			default:
				return s.exitError(ctx, ExitAborted, nil, "user aborted")
			}
		}
	}
	for _, r := range recipients {
		keys, err := s.gpg.FindPublicKeys(ctx, r)
		if err != nil {
			out.Cyan(ctx, "Failed to list public key '%s': %s", r, err)
			continue
		}
		keys = keys.UseableKeys()
		if len(keys) < 1 {
			out.Cyan(ctx, "Warning: No matching valid key found. If the key is in your keyring you may need to validate it.")
			out.Cyan(ctx, "If this is your key: gpg --edit-key %s; trust (set to ultimate); quit", r)
			out.Cyan(ctx, "If this is not your key: gpg --edit-key %s; lsign; save; quit", r)
			continue
		}

		if !s.AskForConfirmation(ctx, fmt.Sprintf("Do you want to add '%s' as an recipient?", keys[0].OneLine())) {
			continue
		}

		if err := s.Store.AddRecipient(ctxutil.WithNoConfirm(ctx, true), store, keys[0].Fingerprint); err != nil {
			return s.exitError(ctx, ExitRecipients, err, "failed to add recipient '%s': %s", r, err)
		}
		added++
	}
	if added < 1 {
		return s.exitError(ctx, ExitUnknown, nil, "no key added")
	}

	out.Green(ctx, "Added %d recipients\n", added)
	return nil
}

// RecipientsRemove removes recipients
func (s *Action) RecipientsRemove(ctx context.Context, c *cli.Context) error {
	store := c.String("store")

	recipients := []string(c.Args())
	if len(recipients) < 1 {
		ids := s.Store.ListRecipients(ctx, store)
		choices := make([]string, 0, len(ids))
		kl, err := s.gpg.FindPublicKeys(ctx, ids...)
		if err == nil && kl != nil {
			for _, id := range ids {
				if key, err := kl.FindKey(id); err == nil {
					choices = append(choices, key.OneLine())
					continue
				}
				choices = append(choices, id)
			}
		}
		if len(choices) > 0 {
			act, sel := termwiz.GetSelection(ctx, "Remove recipient -", "<↑/↓> to change the selection, <→> to remove this recipient, <ESC> to quit", choices)
			switch act {
			case "show":
				recipients = []string{choices[sel]}
			default:
				return s.exitError(ctx, ExitAborted, nil, "user aborted")
			}
		}
	}

	removed := 0
	for _, r := range recipients {
		kl, err := s.gpg.FindPrivateKeys(ctx, r)
		if err == nil {
			if len(kl) > 0 {
				if !s.AskForConfirmation(ctx, fmt.Sprintf("Do you want to remove yourself (%s) from the recipients?", r)) {
					continue
				}
			}
		}
		if err := s.Store.RemoveRecipient(ctxutil.WithNoConfirm(ctx, true), store, strings.TrimPrefix(r, "0x")); err != nil {
			return s.exitError(ctx, ExitRecipients, err, "failed to remove recipient '%s': %s", r, err)
		}
		fmt.Printf(removalWarning, r)
		removed++
	}

	fmt.Printf("Removed %d recipients\n", removed)
	return nil
}
