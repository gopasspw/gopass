package password

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/tree"
)

const (
	// TemplateFile is the name of a pass template
	TemplateFile = ".pass-template"
)

// LookupTemplate will lookup and return a template
func (r *RootStore) LookupTemplate(name string) ([]byte, bool) {
	store := r.getStore(name)
	return store.LookupTemplate(strings.TrimPrefix(name, store.alias))
}

// LookupTemplate will lookup and return a template
func (s *Store) LookupTemplate(name string) ([]byte, bool) {
	for {
		if !strings.Contains(name, string(filepath.Separator)) {
			return []byte{}, false
		}
		name = filepath.Dir(name)
		tpl := filepath.Join(s.path, name, TemplateFile)
		if fsutil.IsFile(tpl) {
			if content, err := ioutil.ReadFile(tpl); err == nil {
				return content, true
			}
		}
	}
}

// TemplateTree returns a tree of all templates
func (r *RootStore) TemplateTree() (*tree.Folder, error) {
	root := tree.New("gopass")
	mps := r.mountPoints()
	sort.Sort(sort.Reverse(byLen(mps)))
	for _, alias := range mps {
		substore := r.mounts[alias]
		if substore == nil {
			continue
		}
		if err := root.AddMount(alias, substore.path); err != nil {
			return nil, fmt.Errorf("failed to add mount: %s", err)
		}
		for _, t := range substore.ListTemplates(alias) {
			// TODO(dschulz) maybe: if err := root.AddFile(t); err != nil {
			if err := root.AddFile(alias + "/" + t); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, t := range r.store.ListTemplates("") {
		if err := root.AddFile(t); err != nil {
			fmt.Println(err)
		}
	}

	return root, nil
}

func mkTemplateStoreWalkerFunc(alias, folder string, fn func(...string)) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != folder {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() != TemplateFile {
			return nil
		}
		if path == folder {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		s := strings.TrimPrefix(path, folder+"/")
		s = strings.TrimSuffix(s, "/"+TemplateFile)
		if alias != "" {
			s = alias + "/" + s
		}
		fn(s)
		return nil
	}
}

// ListTemplates will list all templates in this store
func (s *Store) ListTemplates(prefix string) []string {
	lst := make([]string, 0, 10)
	addFunc := func(in ...string) {
		for _, s := range in {
			lst = append(lst, s)
		}
	}

	if err := filepath.Walk(s.path, mkTemplateStoreWalkerFunc(prefix, s.path, addFunc)); err != nil {
		fmt.Printf("Failed to list templates: %s\n", err)
	}

	return lst
}

// templatefile returns the name of the given template on disk
func (s *Store) templatefile(name string) string {
	return filepath.Join(s.path, name, TemplateFile)
}

// HasTemplate returns true if the template exists
func (r *RootStore) HasTemplate(name string) bool {
	store := r.getStore(name)
	return store.HasTemplate(strings.TrimPrefix(name, store.alias))
}

// HasTemplate returns true if the template exists
func (s *Store) HasTemplate(name string) bool {
	return fsutil.IsFile(s.templatefile(name))
}

// GetTemplate will return the content of the named template
func (r *RootStore) GetTemplate(name string) ([]byte, error) {
	store := r.getStore(name)
	return store.GetTemplate(strings.TrimPrefix(name, store.alias))
}

// GetTemplate will return the content of the named template
func (s *Store) GetTemplate(name string) ([]byte, error) {
	return ioutil.ReadFile(s.templatefile(name))
}

// SetTemplate will (over)write the content to the template file
func (r *RootStore) SetTemplate(name string, content []byte) error {
	store := r.getStore(name)
	return store.SetTemplate(strings.TrimPrefix(name, store.alias), content)
}

// SetTemplate will (over)write the content to the template file
func (s *Store) SetTemplate(name string, content []byte) error {
	return ioutil.WriteFile(s.templatefile(name), content, 0600)
}

// RemoveTemplate will delete the named template if it exists
func (r *RootStore) RemoveTemplate(name string) error {
	store := r.getStore(name)
	return store.RemoveTemplate(strings.TrimPrefix(name, store.alias))
}

// RemoveTemplate will delete the named template if it exists
func (s *Store) RemoveTemplate(name string) error {
	t := s.templatefile(name)
	if !fsutil.IsFile(t) {
		return fmt.Errorf("template not found")
	}

	return os.Remove(t)
}
