package root

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/leaf"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/debug"
)

// ListRecipients lists all recipients for the given store.
func (r *Store) ListRecipients(ctx context.Context, store string) []string {
	sub, _ := r.getStore(store)

	return sub.Recipients(ctx)
}

// CheckRecipients checks all current recipients to make sure that they are
// valid, e.g. not expired.
func (r *Store) CheckRecipients(ctx context.Context, store string) error {
	sub, _ := r.getStore(store)

	return sub.CheckRecipients(ctx)
}

// AddRecipient adds a single recipient to the given store.
func (r *Store) AddRecipient(ctx context.Context, store, rec string) error {
	sub, _ := r.getStore(store)

	return sub.AddRecipient(ctx, rec)
}

// RemoveRecipient removes a single recipient from the given store.
func (r *Store) RemoveRecipient(ctx context.Context, store, rec string) error {
	sub, _ := r.getStore(store)

	return sub.RemoveRecipient(ctx, rec)
}

func (r *Store) addRecipient(ctx context.Context, prefix string, root *tree.Root, recp string, pretty bool) error {
	sub, _ := r.getStore(prefix)
	key := recp

	if pretty {
		key = fmt.Sprintf("%s (missing public key)", recp)

		if v := sub.Crypto().FormatKey(ctx, recp, ""); v != "" {
			key = v
			if !strings.HasPrefix(v, recp) {
				key = recp + " => " + v
			}
			debug.Log("formated (FormatKey) %s as %s", recp, key)
		}
	}

	// workaround to keep key names from breaking the folder structure.
	// A proper fix should change tree.AddFile to take a path and file name
	// (which could then contain slashes).
	key = strings.ReplaceAll(key, "/", "")

	debug.Log("adding %q to the tree", key)

	return r.addFormattedRecipient(root, prefix, recp, key)
}

func (r *Store) addFormattedRecipient(root *tree.Root, prefix, recp, key string) error {
	key = strings.ReplaceAll(key, "/", "")
	debug.Log("adding %q to the tree", key)

	return root.AddFile(prefix+key, "gopass/recipient")
}

func (r *Store) addRecipients(ctx context.Context, prefix string, root *tree.Root, recps []string, pretty bool) error {
	if !pretty {
		for _, recp := range recps {
			if err := r.addRecipient(ctx, prefix, root, recp, false); err != nil {
				return err
			}
		}

		return nil
	}

	sub, _ := r.getStore(prefix)
	formatted := sub.Crypto().FormatKeys(ctx, recps)
	for _, recp := range recps {
		key := formatted[recp]
		if key == "" {
			if err := r.addRecipient(ctx, prefix, root, recp, true); err != nil {
				return err
			}

			continue
		}
		if !strings.HasPrefix(key, recp) {
			key = recp + " => " + key
		}
		if err := r.addFormattedRecipient(root, prefix, recp, key); err != nil {
			return err
		}
	}

	return nil
}

// ImportMissingPublicKeys import missing public keys in any substore.
func (r *Store) ImportMissingPublicKeys(ctx context.Context) error {
	for alias, sub := range r.mounts {
		if err := sub.ImportMissingPublicKeys(ctx); err != nil {
			out.Errorf(ctx, "[%s] Failed to import missing public keys: %s", alias, err)
		}
	}

	return r.store.ImportMissingPublicKeys(ctx)
}

// SaveRecipients persists the recipients to disk. Only useful if persist keys is
// enabled.
func (r *Store) SaveRecipients(ctx context.Context, ack bool) error {
	for alias, sub := range r.mounts {
		if err := sub.SaveRecipients(ctx, ack); err != nil {
			out.Errorf(ctx, "[%s] Failed to save recipients: %s", alias, err)
		}
	}

	return r.store.SaveRecipients(ctx, ack)
}

// RecipientsTree returns a tree view of all stores' recipients.
func (r *Store) RecipientsTree(ctx context.Context, pretty bool) (*tree.Root, error) {
	root := tree.New("gopass")

	for name, recps := range r.store.RecipientsTree(ctx) {
		if name != "" {
			name += "/"
		}

		debug.Log("Store/Secret: %q -> Recipients: %v", name, recps)

		if err := r.addRecipients(ctx, name, root, recps, pretty); err != nil {
			color.Yellow("Failed to add recipients to tree: %s", err)
		}
	}

	mps := r.MountPoints()
	sort.Sort(store.ByPathLen(mps))

	for _, alias := range mps {
		substore := r.mounts[alias]

		// ignore invalid entries
		if substore == nil {
			continue
		}

		if err := root.AddMount(alias, substore.Path()); err != nil {
			return nil, fmt.Errorf("failed to add mount: %w", err)
		}

		for name, recps := range substore.RecipientsTree(ctx) {
			if name != "" {
				name += "/"
			}

			if err := r.addRecipients(ctx, alias+"/"+name, root, recps, pretty); err != nil {
				debug.Log("Failed to add recipients to tree: %s", err)
			}
		}
	}

	return root, nil
}

// RecipientsTreePrefix returns a recipient tree limited to the store selected
// by prefix. The prefix is still retained in the tree so callers can select a
// further subtree from the result.
func (r *Store) RecipientsTreePrefix(ctx context.Context, pretty bool, prefix string) (*tree.Root, error) {
	sub, _ := r.getStore(prefix)
	storePrefix := sub.Alias()
	recipientTree := sub.RecipientsTree(ctx)

	// Build an unformatted tree first. This lets callers reject a missing
	// prefix without doing any GPG key lookups.
	raw := tree.New("gopass")
	for name, recps := range recipientTree {
		path := storePrefix
		if name != "" {
			path += "/" + name
		}
		if path != "" {
			path += "/"
		}

		if err := r.addRecipients(ctx, path, raw, recps, false); err != nil {
			debug.Log("Failed to add recipients to tree: %s", err)
		}
	}
	if _, err := raw.FindFolder(prefix); err != nil {
		return nil, err
	}

	root := tree.New("gopass")

	for name, recps := range recipientTree {
		path := storePrefix
		if name != "" {
			path += "/" + name
		}
		if path != "" {
			path += "/"
		}

		if err := r.addRecipients(ctx, path, root, recps, pretty); err != nil {
			debug.Log("Failed to add recipients to tree: %s", err)
		}
	}

	return root, nil
}

// CanonicalizeRecipients migrates the given store's .gpg-id to use canonical
// (full-fingerprint) recipient IDs and renames the corresponding .public-keys/
// files to match. See leaf.Store.CanonicalizeRecipients for details.
func (r *Store) CanonicalizeRecipients(ctx context.Context, store string) error {
	sub, _ := r.getStore(store)

	return sub.CanonicalizeRecipients(ctx)
}

// DiagnoseRecipients performs a read-only diagnostic of the recipient list
// for the given store. It returns findings about non-canonical IDs,
// unresolvable recipients, and .public-keys/ availability.
// See leaf.Store.DiagnoseRecipients for details.
func (r *Store) DiagnoseRecipients(ctx context.Context, store string) leaf.RecipientDiagnostics {
	sub, _ := r.getStore(store)

	return sub.DiagnoseRecipients(ctx)
}

// JoinTeam performs post-clone / post-setup processing: imports all
// .public-keys/ into the local keyring, checks decryption, and — if needed
// — exports the user's own key additively. Returns true if the user's key
// was newly exported.
func (r *Store) JoinTeam(ctx context.Context, store string) (bool, error) {
	sub, _ := r.getStore(store)

	return sub.JoinTeam(ctx)
}

// UpdateRecipientKeys re-exports the named recipients' public keys into
// .public-keys/. If no IDs are provided, the current user's own identity
// is used. See leaf.Store.UpdateRecipientKeys for details.
func (r *Store) UpdateRecipientKeys(ctx context.Context, store string, ids []string) error {
	sub, _ := r.getStore(store)

	return sub.UpdateRecipientKeys(ctx, ids)
}
