package root

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/tree"
	"github.com/justwatchcom/gopass/tree/simple"
)

// ListRecipients lists all recipients for the given store
func (r *Store) ListRecipients(store string) []string {
	return r.getStore(store).Recipients()
}

// AddRecipient adds a single recipient to the given store
func (r *Store) AddRecipient(store, rec string) error {
	return r.getStore(store).AddRecipient(rec)
}

// RemoveRecipient removes a single recipient from the given store
func (r *Store) RemoveRecipient(store, rec string) error {
	return r.getStore(store).RemoveRecipient(rec)
}

func (r *Store) addRecipient(prefix string, root tree.Tree, recp string, pretty bool) error {
	key := fmt.Sprintf("%s (missing public key)", recp)
	kl, err := r.gpg.FindPublicKeys(recp)
	if err == nil {
		if len(kl) > 0 {
			if pretty {
				key = kl[0].OneLine()
			} else {
				key = kl[0].Fingerprint
			}
		}
	}
	return root.AddFile(prefix+key, "gopass/recipient")
}

// ImportMissingPublicKeys import missing public keys in any substore
func (r *Store) ImportMissingPublicKeys() error {
	if !r.loadKeys {
		return nil
	}

	for alias, sub := range r.mounts {
		if err := sub.ImportMissingPublicKeys(); err != nil {
			fmt.Println(color.RedString("[%s] Failed to import missing public keys: %s", alias, err))
		}
	}

	return r.store.ImportMissingPublicKeys()
}

// SaveRecipients persists the recipients to disk. Only useful if persist keys is
// enabled
func (r *Store) SaveRecipients() error {
	if !r.persistKeys {
		return nil
	}

	for alias, sub := range r.mounts {
		if err := sub.SaveRecipients(); err != nil {
			fmt.Println(color.RedString("[%s] Failed to save recipients: %s", alias, err))
		}
	}

	return r.store.SaveRecipients()
}

// RecipientsTree returns a tree view of all stores' recipients
func (r *Store) RecipientsTree(pretty bool) (tree.Tree, error) {
	root := simple.New("gopass")

	for _, recp := range r.store.Recipients() {
		if err := r.addRecipient("", root, recp, pretty); err != nil {
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
			return nil, fmt.Errorf("failed to add mount: %s", err)
		}
		for _, recp := range substore.Recipients() {
			if err := r.addRecipient(alias+"/", root, recp, pretty); err != nil {
				color.Yellow("Failed to add recipient to tree %s: %s", recp, err)
			}
		}
	}

	return root, nil
}
