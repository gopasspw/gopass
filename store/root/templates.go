package root

import (
	"context"
	"fmt"
	"sort"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/tree"
	"github.com/justwatchcom/gopass/utils/tree/simple"
	"github.com/pkg/errors"
)

// LookupTemplate will lookup and return a template
func (r *Store) LookupTemplate(ctx context.Context, name string) ([]byte, bool) {
	_, store, name := r.getStore(ctx, name)
	return store.LookupTemplate(name)
}

// TemplateTree returns a tree of all templates
func (r *Store) TemplateTree() (tree.Tree, error) {
	root := simple.New("gopass")

	for _, t := range r.store.ListTemplates("") {
		if err := root.AddFile(t, "gopass/template"); err != nil {
			fmt.Println(err)
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
		for _, t := range substore.ListTemplates(alias) {
			if err := root.AddFile(t, "gopass/template"); err != nil {
				fmt.Println(err)
			}
		}
	}

	return root, nil
}

// HasTemplate returns true if the template exists
func (r *Store) HasTemplate(ctx context.Context, name string) bool {
	_, store, name := r.getStore(ctx, name)
	return store.HasTemplate(name)
}

// GetTemplate will return the content of the named template
func (r *Store) GetTemplate(ctx context.Context, name string) ([]byte, error) {
	_, store, name := r.getStore(ctx, name)
	return store.GetTemplate(name)
}

// SetTemplate will (over)write the content to the template file
func (r *Store) SetTemplate(ctx context.Context, name string, content []byte) error {
	_, store, name := r.getStore(ctx, name)
	return store.SetTemplate(name, content)
}

// RemoveTemplate will delete the named template if it exists
func (r *Store) RemoveTemplate(ctx context.Context, name string) error {
	_, store, name := r.getStore(ctx, name)
	return store.RemoveTemplate(name)
}
