package root

import (
	"context"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/internal/tree/simple"

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
				out.Error(ctx, "Failed to add file %s to tree: %s", f, err)
				continue
			}
		}
	}
	addTplFunc := func(in ...string) {
		for _, f := range in {
			if err := root.AddTemplate(f); err != nil {
				out.Error(ctx, "Failed to add template %s to tree: %s", f, err)
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

// HasSubDirs returns true if the named entity has subdirectories
func (r *Store) HasSubDirs(ctx context.Context, name string) (bool, error) {
	ctx, sub, prefix := r.getStore(ctx, name)
	entries, err := sub.List(ctx, prefix)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if sub.IsDir(ctx, e) {
			return true, nil
		}
	}
	return false, nil
}

// Format will pretty print all entries in this store and all substores
func (r *Store) Format(ctx context.Context, maxDepth int) (string, error) {
	t, err := r.Tree(ctx)
	if err != nil {
		return "", err
	}
	return t.Format(maxDepth), nil
}
