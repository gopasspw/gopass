package root

import (
	"fmt"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/tree"
	"github.com/justwatchcom/gopass/tree/simple"
)

// LookupTemplate will lookup and return a template
func (r *Store) LookupTemplate(name string) ([]byte, bool) {
	store := r.getStore(name)
	return store.LookupTemplate(strings.TrimPrefix(name, store.Alias()))
}

// TemplateTree returns a tree of all templates
func (r *Store) TemplateTree() (tree.Tree, error) {
	root := simple.New("gopass")
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
		for _, t := range substore.ListTemplates(alias) {
			if err := root.AddFile(t, "gopass/template"); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, t := range r.store.ListTemplates("") {
		if err := root.AddFile(t, "gopass/template"); err != nil {
			fmt.Println(err)
		}
	}

	return root, nil
}

// HasTemplate returns true if the template exists
func (r *Store) HasTemplate(name string) bool {
	store := r.getStore(name)
	return store.HasTemplate(strings.TrimPrefix(name, store.Alias()))
}

// GetTemplate will return the content of the named template
func (r *Store) GetTemplate(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetTemplate(strings.TrimPrefix(name, store.Alias()))
}

// SetTemplate will (over)write the content to the template file
func (r *Store) SetTemplate(name string, content []byte) error {
	store := r.getStore(name)
	return store.SetTemplate(strings.TrimPrefix(name, store.Alias()), content)
}

// RemoveTemplate will delete the named template if it exists
func (r *Store) RemoveTemplate(name string) error {
	store := r.getStore(name)
	return store.RemoveTemplate(strings.TrimPrefix(name, store.Alias()))
}
