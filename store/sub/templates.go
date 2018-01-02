package sub

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/tree"
	"github.com/justwatchcom/gopass/utils/tree/simple"
	"github.com/pkg/errors"
)

const (
	// TemplateFile is the name of a pass template
	TemplateFile = ".pass-template"
)

// LookupTemplate will lookup and return a template
func (s *Store) LookupTemplate(name string) ([]byte, bool) {
	// chop off one path element until we find something
	for {
		l1 := len(name)
		name = filepath.Dir(name)
		if len(name) == l1 {
			break
		}
		tpl := filepath.Join(s.path, name, TemplateFile)
		if fsutil.IsFile(tpl) {
			if content, err := ioutil.ReadFile(tpl); err == nil {
				return content, true
			}
		}
	}
	return []byte{}, false
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
		s := strings.TrimPrefix(path, folder+sep)
		s = strings.TrimSuffix(s, TemplateFile)
		s = strings.TrimSuffix(s, sep)
		if s == "" {
			s = "default"
		}
		if alias != "" {
			s = alias + sep + s
		}
		// make sure to always use forward slashes for internal gopass representation
		s = filepath.ToSlash(s)
		fn(s)
		return nil
	}
}

// ListTemplates will list all templates in this store
func (s *Store) ListTemplates(prefix string) []string {
	lst := make([]string, 0, 10)
	addFunc := func(in ...string) {
		lst = append(lst, in...)
	}

	path, err := filepath.EvalSymlinks(s.path)
	if err != nil {
		return lst
	}
	if err := filepath.Walk(path, mkTemplateStoreWalkerFunc(prefix, path, addFunc)); err != nil {
		fmt.Printf("Failed to list templates: %s\n", err)
	}

	return lst
}

// TemplateTree returns a tree of all templates
func (s *Store) TemplateTree() (tree.Tree, error) {
	root := simple.New("gopass")
	for _, t := range s.ListTemplates("") {
		if err := root.AddFile(t, "gopass/template"); err != nil {
			fmt.Println(err)
		}
	}

	return root, nil
}

// templatefile returns the name of the given template on disk
func (s *Store) templatefile(name string) string {
	return filepath.Join(s.path, name, TemplateFile)
}

// HasTemplate returns true if the template exists
func (s *Store) HasTemplate(name string) bool {
	return fsutil.IsFile(s.templatefile(name))
}

// GetTemplate will return the content of the named template
func (s *Store) GetTemplate(name string) ([]byte, error) {
	return ioutil.ReadFile(s.templatefile(name))
}

// SetTemplate will (over)write the content to the template file
func (s *Store) SetTemplate(name string, content []byte) error {
	tplFile := s.templatefile(name)
	tplDir := filepath.Dir(tplFile)
	if err := os.MkdirAll(tplDir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(tplFile, content, 0600)
}

// RemoveTemplate will delete the named template if it exists
func (s *Store) RemoveTemplate(name string) error {
	t := s.templatefile(name)
	if !fsutil.IsFile(t) {
		return errors.Errorf("template not found")
	}

	return os.Remove(t)
}
