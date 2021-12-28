package action

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gopasspw/gopass/internal/editor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tpl"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/urfave/cli/v2"
)

const (
	templateExample = `{{ .Content }}

# This is an example of the available template operations
# Predefined variables:
# - .Content: The secret payload, usually a generated password
# - .Name: The name of this secret
# - .Path: The path to this secret
# - .Dir: The dir of this secret
#
# Available Template functions:
# - md5sum: e.g. {{ .Content | md5sum }}
# - sha1sum: e.g. {{ .Content | sha1sum }}
# - md5crypt: e.g. {{ .Content | md5crypt }}
# - ssha: e.g. {{ .Content | ssha }}
# - ssha256: e.g. {{ .Content | ssha256 }}
# - ssha512: e.g. {{ .Content | ssha512 }}
# - get "key": e.g. {{ get "path/to/some/other/secret" | md5sum }}
`
)

// TemplatesPrint will pretty-print a tree of templates
func (s *Action) TemplatesPrint(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	t, err := s.Store.TemplateTree(ctx)
	if err != nil {
		return ExitError(ExitList, err, "failed to list templates: %s", err)
	}
	fmt.Fprintln(stdout, t.Format(tree.INF))
	return nil
}

// TemplatePrint will lookup and print a single template
func (s *Action) TemplatePrint(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()

	content, err := s.Store.GetTemplate(ctx, name)
	if err != nil {
		return ExitError(ExitIO, err, "failed to retrieve template: %s", err)
	}

	fmt.Fprintln(stdout, string(content))
	return nil
}

// TemplateEdit will load and existing or new template into an
// editor
func (s *Action) TemplateEdit(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()

	var content []byte
	if s.Store.HasTemplate(ctx, name) {
		var err error
		content, err = s.Store.GetTemplate(ctx, name)
		if err != nil {
			return ExitError(ExitIO, err, "failed to retrieve template: %s", err)
		}
	} else {
		content = []byte(templateExample)
	}

	ed := editor.Path(c)
	nContent, err := editor.Invoke(ctx, ed, content)
	if err != nil {
		return ExitError(ExitUnknown, err, "failed to invoke editor %s: %s", ed, err)
	}

	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) {
		return nil
	}

	return s.Store.SetTemplate(ctx, name, nContent)
}

// TemplateRemove will remove a single template
func (s *Action) TemplateRemove(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	if name == "" {
		return ExitError(ExitUsage, nil, "usage: %s templates remove [name]", s.Name)
	}

	if !s.Store.HasTemplate(ctx, name) {
		return ExitError(ExitNotFound, nil, "template %q not found", name)
	}

	return s.Store.RemoveTemplate(ctx, name)
}

func (s *Action) templatesList(ctx context.Context) []string {
	t, err := s.Store.TemplateTree(ctx)
	if err != nil {
		debug.Log("failed to list templates: %s", err)
		return nil
	}

	return t.List(tree.INF)
}

// TemplatesComplete prints a list of all templates for bash completion
func (s *Action) TemplatesComplete(c *cli.Context) {
	ctx := ctxutil.WithGlobalFlags(c)

	for _, v := range s.templatesList(ctx) {
		fmt.Fprintln(stdout, v)
	}
}

func (s *Action) renderTemplate(ctx context.Context, name string, content []byte) ([]byte, bool) {
	tName, tmpl, found := s.Store.LookupTemplate(ctx, name)
	if !found {
		debug.Log("No template found for %s", name)
		return content, false
	}

	tmplStr := strings.TrimSpace(string(tmpl))
	if tmplStr == "" {
		debug.Log("Skipping empty template %q, for %s", tName, name)
		return content, false
	}

	// load template if it exists
	nc, err := tpl.Execute(ctx, string(tmpl), name, content, s.Store)
	if err != nil {
		fmt.Fprintf(stdout, "failed to execute template %q: %s\n", tName, err)
		return content, false
	}

	out.Printf(ctx, "Note: Using template %s", tName)

	return nc, true
}
