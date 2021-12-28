package action

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/cui"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/urfave/cli/v2"
)

var (
	removalWarning = `


@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@   WARNING: REMOVING A USER WILL NOT REVOKE ACCESS FROM OLD REVISONS!   @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
THE USER %s WILL STILL BE ABLE TO ACCESS ANY OLD COPY OF THE STORE AND
ANY OLD REVISION THEY HAD ACCESS TO.

ANY CREDENTIALS THIS USER HAD ACCESS TO NEED TO BE CONSIDERED COMPROMISED
AND SHOULD BE REVOKED.

This feature is only meant for revoking access to any added or changed
credentials.

`
)

// RecipientsPrint prints all recipients per store
func (s *Action) RecipientsPrint(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	out.Printf(ctx, "Hint: run 'gopass sync' to import any missing public keys")

	t, err := s.Store.RecipientsTree(ctx, true)
	if err != nil {
		return ExitError(ExitList, err, "failed to list recipients: %s", err)
	}

	fmt.Fprintln(stdout, t.Format(tree.INF))
	return nil
}

func (s *Action) recipientsList(ctx context.Context) []string {
	t, err := s.Store.RecipientsTree(ctxutil.WithHidden(ctx, true), false)
	if err != nil {
		debug.Log("failed to list recipients: %s", err)
		return nil
	}

	return t.List(tree.INF)
}

// RecipientsComplete will print a list of recipients for bash
// completion
func (s *Action) RecipientsComplete(c *cli.Context) {
	ctx := ctxutil.WithGlobalFlags(c)
	for _, v := range s.recipientsList(ctx) {
		fmt.Fprintln(stdout, v)
	}
}

// RecipientsAdd adds new recipients
func (s *Action) RecipientsAdd(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	force := c.Bool("force")
	added := 0

	// select store
	if store == "" {
		store = cui.AskForStore(ctx, s.Store)
	}

	crypto := s.Store.Crypto(ctx, store)

	// select recipient
	recipients := []string(c.Args().Slice())
	if len(recipients) < 1 {
		debug.Log("no recipients given, asking for selection")
		r, err := s.recipientsSelectForAdd(ctx, store)
		if err != nil {
			return err
		}
		recipients = r
	}

	debug.Log("adding recipients: %+v", recipients)
	for _, r := range recipients {
		keys, err := crypto.FindRecipients(ctx, r)
		if err != nil {
			out.Printf(ctx, "WARNING: Failed to list public key %q: %s", r, err)
			if !force {
				continue
			}
			keys = []string{r}
		}
		if len(keys) < 1 && !force && crypto.Name() == "gpgcli" {
			out.Printf(ctx, "Warning: No matching valid key found. If the key is in your keyring you may need to validate it.")
			out.Printf(ctx, "If this is your key: gpg --edit-key %s; trust (set to ultimate); quit", r)
			out.Printf(ctx, "If this is not your key: gpg --edit-key %s; lsign; trust; save; quit", r)
			out.Printf(ctx, "You may need to run 'gpg --update-trustdb' afterwards")
			continue
		}

		recp := r
		debug.Log("found recipients for %q: %+v", r, keys)

		if !termio.AskForConfirmation(ctx, fmt.Sprintf("Do you want to add %q (key %q) as a recipient to the store %q?", crypto.FormatKey(ctx, recp, ""), recp, store)) {
			continue
		}

		if err := s.Store.AddRecipient(ctx, store, recp); err != nil {
			return ExitError(ExitRecipients, err, "failed to add recipient %q: %s", r, err)
		}
		added++
	}
	if added < 1 {
		return ExitError(ExitUnknown, nil, "no key added")
	}

	out.Printf(ctx, "\nAdded %d recipients", added)
	out.Printf(ctx, "You need to run 'gopass sync' to push these changes")
	return nil
}

// RecipientsRemove removes recipients
func (s *Action) RecipientsRemove(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	store := c.String("store")
	force := c.Bool("force")
	removed := 0

	// select store
	if store == "" {
		store = cui.AskForStore(ctx, s.Store)
	}

	crypto := s.Store.Crypto(ctx, store)

	// select recipient
	recipients := []string(c.Args().Slice())
	if len(recipients) < 1 {
		rs, err := s.recipientsSelectForRemoval(ctx, store)
		if err != nil {
			return err
		}
		recipients = rs
	}

	for _, r := range recipients {
		kl, err := crypto.FindIdentities(ctx, r)
		if err == nil {
			if len(kl) > 0 {
				if !termio.AskForConfirmation(ctx, fmt.Sprintf("Do you want to remove yourself (%s) from the recipients?", r)) {
					continue
				}
			}
		}

		keys, err := crypto.FindRecipients(ctx, r)
		if err != nil {
			out.Printf(ctx, "WARNING: Failed to list public key %q: %s", r, err)
			out.Printf(ctx, "Hint: You can use `--force` to remove unknown keys.")
			if !force {
				continue
			}
			keys = []string{r}
		}
		if len(keys) < 1 && !force {
			out.Printf(ctx, "Warning: No matching valid key found. If the key is in your keyring you may need to validate it.")
			out.Printf(ctx, "If this is your key: gpg --edit-key %s; trust (set to ultimate); quit", r)
			out.Printf(ctx, "If this is not your key: gpg --edit-key %s; lsign; trust; save; quit", r)
			out.Printf(ctx, "You may need to run 'gpg --update-trustdb' afterwards")
			continue
		}

		recp := r
		if len(keys) > 0 {
			recp = crypto.Fingerprint(ctx, keys[0])
		}

		if err := s.Store.RemoveRecipient(ctx, store, recp); err != nil {
			return ExitError(ExitRecipients, err, "failed to remove recipient %q: %s", recp, err)
		}
		fmt.Fprintf(stdout, removalWarning, r)
		removed++
	}
	if removed < 1 {
		return ExitError(ExitUnknown, nil, "no key removed")
	}

	out.Printf(ctx, "\nRemoved %d recipients", removed)
	out.Printf(ctx, "You need to run 'gopass sync' to push these changes")
	return nil
}

func (s *Action) recipientsSelectForRemoval(ctx context.Context, store string) ([]string, error) {
	crypto := s.Store.Crypto(ctx, store)

	ids := s.Store.ListRecipients(ctx, store)
	choices := make([]string, 0, len(ids))
	for _, id := range ids {
		choices = append(choices, crypto.FormatKey(ctx, id, ""))
	}
	if len(choices) < 1 {
		return nil, nil
	}

	act, sel := cui.GetSelection(ctx, "Remove recipient -", choices)
	switch act {
	case "default":
		fallthrough
	case "show":
		return []string{ids[sel]}, nil
	default:
		return nil, ExitError(ExitAborted, nil, "user aborted")
	}
}

func (s *Action) recipientsSelectForAdd(ctx context.Context, store string) ([]string, error) {
	crypto := s.Store.Crypto(ctx, store)

	choices := []string{}
	kl, _ := crypto.FindRecipients(ctx)
	for _, key := range kl {
		choices = append(choices, crypto.FormatKey(ctx, key, ""))
	}
	if len(choices) < 1 {
		return nil, nil
	}

	act, sel := cui.GetSelection(ctx, "Add Recipient -", choices)
	switch act {
	case "default":
		fallthrough
	case "show":
		return []string{kl[sel]}, nil
	default:
		return nil, ExitError(ExitAborted, nil, "user aborted")
	}
}
