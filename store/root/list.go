package root

import (
	"fmt"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/tree"
	"github.com/justwatchcom/gopass/utils/tree/simple"
	"github.com/pkg/errors"
)

// List will return a flattened list of all tree entries
func (r *Store) List(maxDepth int) ([]string, error) {
	t, err := r.Tree()
	if err != nil {
		return []string{}, err
	}
	return t.List(maxDepth), nil
}

// Tree returns the tree representation of the entries
func (r *Store) Tree() (tree.Tree, error) {
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
				fmt.Printf("Failed to add file %s to tree: %s\n", f, err)
				continue
			}
		}
	}
	addTplFunc := func(in ...string) {
		for _, f := range in {
			if err := root.AddTemplate(f); err != nil {
				fmt.Printf("Failed to add template %s to tree: %s\n", f, err)
				continue
			}
		}
	}

	sf, err := r.store.List("")
	if err != nil {
		return nil, err
	}
	addFileFunc(sf...)
	addTplFunc(r.store.ListTemplates("")...)

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
		sf, err := substore.List(alias)
		if err != nil {
			return nil, errors.Errorf("failed to add file: %s", err)
		}
		addFileFunc(sf...)
		addTplFunc(substore.ListTemplates(alias)...)
	}

	return root, nil
}

// Format will pretty print all entries in this store and all substores
func (r *Store) Format(maxDepth int) (string, error) {
	t, err := r.Tree()
	if err != nil {
		return "", err
	}
	return t.Format(maxDepth), nil
}
