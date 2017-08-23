package action

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
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
func (s *Action) RecipientsPrint(c *cli.Context) error {
	if err := s.Store.ImportMissingPublicKeys(); err != nil {
		fmt.Println(color.RedString("Failed to import missing public keys: %s", err))
	}

	if err := s.Store.SaveRecipients(); err != nil {
		fmt.Println(color.RedString("Failed to export missing public keys: %s", err))
	}

	tree, err := s.Store.RecipientsTree(true)
	if err != nil {
		return s.exitError(ExitList, err, "failed to list recipients: %s", err)
	}

	fmt.Println(tree.Format(0))
	return nil
}

// RecipientsComplete will print a list of recipients for bash
// completion
func (s *Action) RecipientsComplete(*cli.Context) {
	tree, err := s.Store.RecipientsTree(false)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range tree.List(0) {
		fmt.Println(v)
	}
}

// RecipientsAdd adds new recipients
func (s *Action) RecipientsAdd(c *cli.Context) error {
	store := c.String("store")
	added := 0
	for _, r := range c.Args() {
		keys, err := s.gpg.FindPublicKeys(r)
		if err != nil {
			fmt.Println(color.CyanString("Failed to list public key '%s': %s", r, err))
			continue
		}
		keys = keys.UseableKeys()
		if len(keys) < 1 {
			fmt.Println(color.CyanString("Warning: No matching valid key found. If the key is in your keyring you may need to validate it."))
			fmt.Println(color.CyanString("If this is your key: gpg --edit-key %s; trust (set to ultimate); quit", r))
			fmt.Println(color.CyanString("If this is not your key: gpg --edit-key %s; lsign; save; quit", r))
			continue
		}

		if !s.askForConfirmation(fmt.Sprintf("Do you want to add '%s' as an recipient?", keys[0].OneLine())) {
			continue
		}

		if err := s.Store.AddRecipient(store, keys[0].Fingerprint); err != nil {
			return s.exitError(ExitRecipients, err, "failed to add recipient '%s': %s", r, err)
		}
		added++
	}
	if added < 1 {
		return s.exitError(ExitUnknown, nil, "no key added")
	}

	fmt.Println(color.GreenString("Added %d recipients\n", added))
	return nil
}

// RecipientsRemove removes recipients
func (s *Action) RecipientsRemove(c *cli.Context) error {
	store := c.String("store")
	removed := 0
	for _, r := range c.Args() {
		kl, err := s.gpg.FindPrivateKeys(r)
		if err == nil {
			if len(kl) > 0 {
				if !s.askForConfirmation(fmt.Sprintf("Do you want to remove yourself (%s) from the recipients?", r)) {
					continue
				}
			}
		}
		if err := s.Store.RemoveRecipient(store, strings.TrimPrefix(r, "0x")); err != nil {
			return s.exitError(ExitRecipients, err, "failed to remove recipient '%s': %s", r, err)
		}
		fmt.Printf(removalWarning, r)
		removed++
	}

	fmt.Printf("Removed %d recipients\n", removed)
	return nil
}
