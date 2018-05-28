package root

import (
	"context"
	"sort"

	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store"
	"github.com/gopasspw/gopass/pkg/tree"
	"github.com/gopasspw/gopass/pkg/tree/simple"

	"github.com/pkg/errors"
)

// LookupTemplate will lookup and return a template
func (r *Store) LookupTemplate(ctx context.Context, name string) ([]byte, bool) {
	_, store, name := r.getStore(ctx, name)
	return store.LookupTemplate(ctx, name)
}

// TemplateTree returns a tree of all templates
func (r *Store) TemplateTree(ctx context.Context) (tree.Tree, error) {
	root := simple.New("gopass")

	for _, t := range r.store.ListTemplates(ctx, "") {
		if err := root.AddFile(t, "gopass/template"); err != nil {
			out.Red(ctx, "Failed to add file to tree: %s", err)
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
		for _, t := range substore.ListTemplates(ctx, alias) {
			if err := root.AddFile(t, "gopass/template"); err != nil {
				out.Red(ctx, "Failed to add file to tree: %s", err)
			}
		}
	}

	return root, nil
}

// HasTemplate returns true if the template exists
func (r *Store) HasTemplate(ctx context.Context, name string) bool {
	_, store, name := r.getStore(ctx, name)
	return store.HasTemplate(ctx, name)
}

// GetTemplate will return the content of the named template
func (r *Store) GetTemplate(ctx context.Context, name string) ([]byte, error) {
	_, store, name := r.getStore(ctx, name)
	return store.GetTemplate(ctx, name)
}

// SetTemplate will (over)write the content to the template file
func (r *Store) SetTemplate(ctx context.Context, name string, content []byte) error {
	_, store, name := r.getStore(ctx, name)
	return store.SetTemplate(ctx, name, content)
}

// RemoveTemplate will delete the named template if it exists
func (r *Store) RemoveTemplate(ctx context.Context, name string) error {
	_, store, name := r.getStore(ctx, name)
	return store.RemoveTemplate(ctx, name)
}
