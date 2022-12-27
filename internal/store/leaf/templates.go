package leaf

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	// TemplateFile is the name of a pass template.
	TemplateFile = ".pass-template"
)

// LookupTemplate will lookup and return a template.
func (s *Store) LookupTemplate(ctx context.Context, name string) (string, []byte, bool) {
	oName := name
	// go upwards in the directory tree until we find a template
	// by chopping off one path element by one.
	for {
		l1 := len(name)
		name = filepath.Dir(name)

		if len(name) == l1 {
			break
		}

		tpl := filepath.Join(name, TemplateFile)

		if s.storage.Exists(ctx, tpl) {
			if content, err := s.storage.Get(ctx, tpl); err == nil {
				debug.Log("Found template %q for %q", tpl, oName)

				return tpl, content, true
			}
		}
	}

	return "", []byte{}, false
}

// ListTemplates will list all templates in this store.
func (s *Store) ListTemplates(ctx context.Context, prefix string) []string {
	lst, err := s.storage.List(ctx, "")
	if err != nil {
		debug.Log("failed to list templates: %s", err)

		return nil
	}

	tpls := make(map[string]struct{}, len(lst))

	for _, path := range lst {
		if !strings.HasSuffix(path, TemplateFile) {
			continue
		}

		path = strings.TrimSuffix(path, Sep+TemplateFile)

		if prefix != "" {
			path = prefix + Sep + path
		}

		tpls[path] = struct{}{}
	}

	out := make([]string, 0, len(tpls))

	for k := range tpls {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}

// TemplateTree returns a tree of all templates.
func (s *Store) TemplateTree(ctx context.Context) *tree.Root {
	root := tree.New("gopass")

	for _, t := range s.ListTemplates(ctx, "") {
		if err := root.AddFile(t, "gopass/template"); err != nil {
			out.Errorf(ctx, "Failed to add template: %s", err)
		}
	}

	return root
}

// templatefile returns the name of the given template on disk.
func (s *Store) templatefile(name string) string {
	return strings.TrimPrefix(filepath.Join(name, TemplateFile), string(filepath.Separator))
}

// HasTemplate returns true if the template exists.
func (s *Store) HasTemplate(ctx context.Context, name string) bool {
	return s.storage.Exists(ctx, s.templatefile(name))
}

// GetTemplate will return the content of the named template.
func (s *Store) GetTemplate(ctx context.Context, name string) ([]byte, error) {
	return s.storage.Get(ctx, s.templatefile(name))
}

// SetTemplate will (over)write the content to the template file.
func (s *Store) SetTemplate(ctx context.Context, name string, content []byte) error {
	p := s.templatefile(name)

	if err := s.storage.Set(ctx, p, content); err != nil {
		if errors.Is(err, store.ErrMeaninglessWrite) {
			return nil
		}
		return fmt.Errorf("failed to write template: %w", err)
	}

	if err := s.storage.Add(ctx, p); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}

		return fmt.Errorf("failed to add %q to git: %w", p, err)
	}

	if !ctxutil.IsGitCommit(ctx) {
		return nil
	}

	return s.gitCommitAndPush(ctx, name)
}

// RemoveTemplate will delete the named template if it exists.
func (s *Store) RemoveTemplate(ctx context.Context, name string) error {
	p := s.templatefile(name)

	if err := s.storage.Delete(ctx, p); err != nil {
		return fmt.Errorf("failed to remote template: %w", err)
	}

	if err := s.storage.Add(ctx, p); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			return nil
		}

		return fmt.Errorf("failed to add %q to git: %w", p, err)
	}

	if !ctxutil.IsGitCommit(ctx) {
		return nil
	}

	return s.gitCommitAndPush(ctx, name)
}
