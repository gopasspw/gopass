package root

import (
	"fmt"
	"sort"

	"github.com/justwatchcom/gopass/gpg"
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

// RecipientsTree returns a tree view of all stores' recipients
func (r *Store) RecipientsTree(pretty bool) (tree.Tree, error) {
	root := simple.New("gopass")
	mps := r.mountPoints()
	sort.Sort(sort.Reverse(byLen(mps)))
	for _, alias := range mps {
		substore := r.mounts[alias]
		if substore == nil {
			continue
		}
		if err := root.AddMount(alias, substore.Path()); err != nil {
			return nil, fmt.Errorf("failed to add mount: %s", err)
		}
		for _, r := range substore.Recipients() {
			key := fmt.Sprintf("%s (missing public key)", r)
			kl, err := gpg.ListPublicKeys(r)
			if err == nil {
				if len(kl) > 0 {
					if pretty {
						key = kl[0].OneLine()
					} else {
						key = kl[0].Fingerprint
					}
				}
			}
			if err := root.AddFile(alias+"/"+key, "gopass/recipient"); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, r := range r.store.Recipients() {
		kl, err := gpg.ListPublicKeys(r)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(kl) < 1 {
			fmt.Println("key not found", r)
			continue
		}
		key := kl[0].Fingerprint
		if pretty {
			key = kl[0].OneLine()
		}
		if err := root.AddFile(key, "gopass/recipient"); err != nil {
			fmt.Println(err)
		}
	}
	return root, nil
}
