package root

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/debug"
)

// LookupTemplate will lookup and return a template
func (r *Store) LookupTemplate(ctx context.Context, name string) (string, []byte, bool) {
	oName := name
	store, name := r.getStore(name)
	tName, content, found := store.LookupTemplate(ctx, name)
	tName = filepath.Join(r.MountPoint(oName), tName)
	return tName, content, found
}

// TemplateTree returns a tree of all templates
func (r *Store) TemplateTree(ctx context.Context) (*tree.Root, error) {
	root := tree.New("gopass")

	for _, t := range r.store.ListTemplates(ctx, "") {
		debug.Log("[<root>] Adding template %s", t)
		if err := root.AddFile(t, "gopass/template"); err != nil {
			out.Error(ctx, "Failed to add file to tree: %s", err)
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
			return nil, fmt.Errorf("failed to add mount: %w", err)
		}
		for _, t := range substore.ListTemplates(ctx, alias) {
			debug.Log("[%s] Adding template %s", alias, t)
			if err := root.AddFile(t, "gopass/template"); err != nil {
				out.Error(ctx, "Failed to add file to tree: %s", err)
			}
		}
	}

	return root, nil
}

// HasTemplate returns true if the template exists
func (r *Store) HasTemplate(ctx context.Context, name string) bool {
	store, name := r.getStore(name)
	return store.HasTemplate(ctx, name)
}

// GetTemplate will return the content of the named template
func (r *Store) GetTemplate(ctx context.Context, name string) ([]byte, error) {
	store, name := r.getStore(name)
	return store.GetTemplate(ctx, name)
}

// SetTemplate will (over)write the content to the template file
func (r *Store) SetTemplate(ctx context.Context, name string, content []byte) error {
	store, name := r.getStore(name)
	return store.SetTemplate(ctx, name, content)
}

// RemoveTemplate will delete the named template if it exists
func (r *Store) RemoveTemplate(ctx context.Context, name string) error {
	store, name := r.getStore(name)
	return store.RemoveTemplate(ctx, name)
}
