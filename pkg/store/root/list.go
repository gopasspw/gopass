package root

import (
	"context"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/justwatchcom/gopass/pkg/tree"
	"github.com/justwatchcom/gopass/pkg/tree/simple"
	"github.com/pkg/errors"
)

// List will return a flattened list of all tree entries
func (r *Store) List(ctx context.Context, maxDepth int) ([]string, error) {
	t, err := r.Tree(ctx)
	if err != nil {
		return []string{}, err
	}
	return t.List(maxDepth), nil
}

// Tree returns the tree representation of the entries
func (r *Store) Tree(ctx context.Context) (tree.Tree, error) {
	root := simple.New("gopass")
	addFileFunc := func(in ...string) {
		for _, f := range in {
			var ct string
			switch {
			case strings.HasSuffix(f, ".b64"):
				ct = "application/octet-stream"
			case strings.HasSuffix(f, ".yml"):
				ct = "text/yaml"
			case strings.HasSuffix(f, ".yaml"):
				ct = "text/yaml"
			default:
				ct = "text/plain"
			}
			if err := root.AddFile(f, ct); err != nil {
				out.Red(ctx, "Failed to add file %s to tree: %s", f, err)
				continue
			}
		}
	}
	addTplFunc := func(in ...string) {
		for _, f := range in {
			if err := root.AddTemplate(f); err != nil {
				out.Red(ctx, "Failed to add template %s to tree: %s", f, err)
				continue
			}
		}
	}

	sf, err := r.store.List(ctx, "")
	if err != nil {
		return nil, err
	}
	addFileFunc(sf...)
	addTplFunc(r.store.ListTemplates(ctx, "")...)

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
		sf, err := substore.List(ctx, "")
		if err != nil {
			return nil, errors.Errorf("failed to add file: %s", err)
		}
		addFileFunc(sf...)
		addTplFunc(substore.ListTemplates(ctx, alias)...)
	}

	return root, nil
}

// Format will pretty print all entries in this store and all substores
func (r *Store) Format(ctx context.Context, maxDepth int) (string, error) {
	t, err := r.Tree(ctx)
	if err != nil {
		return "", err
	}
	return t.Format(maxDepth), nil
}
