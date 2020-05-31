package root

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/tree"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// ListRecipients lists all recipients for the given store
func (r *Store) ListRecipients(ctx context.Context, store string) []string {
	ctx, sub, _ := r.getStore(ctx, store)
	return sub.Recipients(ctx)
}

// AddRecipient adds a single recipient to the given store
func (r *Store) AddRecipient(ctx context.Context, store, rec string) error {
	ctx, sub, _ := r.getStore(ctx, store)
	return sub.AddRecipient(ctx, rec)
}

// RemoveRecipient removes a single recipient from the given store
func (r *Store) RemoveRecipient(ctx context.Context, store, rec string) error {
	ctx, sub, _ := r.getStore(ctx, store)
	return sub.RemoveRecipient(ctx, rec)
}

func (r *Store) addRecipient(ctx context.Context, prefix string, root *tree.Root, recp string, pretty bool) error {
	ctx, sub, _ := r.getStore(ctx, prefix)
	key := fmt.Sprintf("%s (missing public key)", recp)
	kl, err := sub.Crypto().FindRecipients(ctx, recp)
	if err == nil {
		if len(kl) > 0 {
			if pretty {
				key = sub.Crypto().FormatKey(ctx, kl[0], "")
			} else {
				key = kl[0]
			}
		}
	}
	// workaround to keep key names from breaking the folder structure.
	// A proper fix should change tree.AddFile to take a path and file name
	// (which could then contain slashes).
	key = strings.Replace(key, "/", "", -1)

	return root.AddFile(prefix+key, "gopass/recipient")
}

// ImportMissingPublicKeys import missing public keys in any substore
func (r *Store) ImportMissingPublicKeys(ctx context.Context) error {
	for alias, sub := range r.mounts {
		if err := sub.ImportMissingPublicKeys(ctx); err != nil {
			out.Error(ctx, "[%s] Failed to import missing public keys: %s", alias, err)
		}
	}

	return r.store.ImportMissingPublicKeys(ctx)
}

// SaveRecipients persists the recipients to disk. Only useful if persist keys is
// enabled
func (r *Store) SaveRecipients(ctx context.Context) error {
	for alias, sub := range r.mounts {
		if err := sub.SaveRecipients(ctx); err != nil {
			out.Error(ctx, "[%s] Failed to save recipients: %s", alias, err)
		}
	}

	return r.store.SaveRecipients(ctx)
}

// RecipientsTree returns a tree view of all stores' recipients
func (r *Store) RecipientsTree(ctx context.Context, pretty bool) (*tree.Root, error) {
	root := tree.New("gopass")

	for _, recp := range r.store.Recipients(ctx) {
		if err := r.addRecipient(ctx, "", root, recp, pretty); err != nil {
			color.Yellow("Failed to add recipient to tree %s: %s", recp, err)
		}
	}

	mps := r.MountPoints()
	sort.Sort(store.ByPathLen(mps))
	for _, alias := range mps {
		substore := r.mounts[alias]
		if substore == nil {
			continue
		}
		if err := root.AddMount(alias, substore.Path()); err != nil {
			return nil, errors.Errorf("failed to add mount: %s", err)
		}
		for _, recp := range substore.Recipients(ctx) {
			if err := r.addRecipient(ctx, alias+"/", root, recp, pretty); err != nil {
				debug.Log("Failed to add recipient to tree %s: %s", recp, err)
			}
		}
	}

	return root, nil
}
