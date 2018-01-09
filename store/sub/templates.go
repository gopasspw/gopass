package sub

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/justwatchcom/gopass/utils/tree"
	"github.com/justwatchcom/gopass/utils/tree/simple"
	"github.com/pkg/errors"
)

const (
	// TemplateFile is the name of a pass template
	TemplateFile = ".pass-template"
)

// LookupTemplate will lookup and return a template
func (s *Store) LookupTemplate(ctx context.Context, name string) ([]byte, bool) {
	// chop off one path element until we find something
	for {
		l1 := len(name)
		name = filepath.Dir(name)
		if len(name) == l1 {
			break
		}
		tpl := filepath.Join(name, TemplateFile)
		if s.store.Exists(ctx, tpl) {
			if content, err := s.store.Get(ctx, tpl); err == nil {
				return content, true
			}
		}
	}
	return []byte{}, false
}

// ListTemplates will list all templates in this store
func (s *Store) ListTemplates(ctx context.Context, prefix string) []string {
	lst, err := s.store.List(ctx, prefix)
	if err != nil {
		out.Debug(ctx, "failed to list templates: %s", err)
		return nil
	}
	tpls := make(map[string]struct{}, len(lst))
	for _, path := range lst {
		if !strings.HasSuffix(path, TemplateFile) {
			continue
		}
		path = strings.TrimSuffix(path, sep+TemplateFile)
		tpls[path] = struct{}{}
	}
	out := make([]string, 0, len(tpls))
	for k := range tpls {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TemplateTree returns a tree of all templates
func (s *Store) TemplateTree(ctx context.Context) (tree.Tree, error) {
	root := simple.New("gopass")
	for _, t := range s.ListTemplates(ctx, "") {
		if err := root.AddFile(t, "gopass/template"); err != nil {
			out.Red(ctx, "Failed to add template: %s", err)
		}
	}

	return root, nil
}

// templatefile returns the name of the given template on disk
func (s *Store) templatefile(name string) string {
	return filepath.Join(name, TemplateFile)
}

// HasTemplate returns true if the template exists
func (s *Store) HasTemplate(ctx context.Context, name string) bool {
	return s.store.Exists(ctx, s.templatefile(name))
}

// GetTemplate will return the content of the named template
func (s *Store) GetTemplate(ctx context.Context, name string) ([]byte, error) {
	return s.store.Get(ctx, s.templatefile(name))
}

// SetTemplate will (over)write the content to the template file
func (s *Store) SetTemplate(ctx context.Context, name string, content []byte) error {
	p := s.templatefile(name)

	if err := s.store.Set(ctx, p, content); err != nil {
		return errors.Wrapf(err, "failed to write template")
	}

	if err := s.sync.Add(ctx, p); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", p)
	}

	if !ctxutil.IsGitCommit(ctx) {
		return nil
	}

	return s.gitCommitAndPush(ctx, name)
}

// RemoveTemplate will delete the named template if it exists
func (s *Store) RemoveTemplate(ctx context.Context, name string) error {
	p := s.templatefile(name)

	if err := s.store.Delete(ctx, p); err != nil {
		return errors.Wrapf(err, "failed to remote template")
	}

	if err := s.sync.Add(ctx, p); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		return errors.Wrapf(err, "failed to add '%s' to git", p)
	}

	if !ctxutil.IsGitCommit(ctx) {
		return nil
	}

	return s.gitCommitAndPush(ctx, name)
}
