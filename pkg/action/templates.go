package action

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/editor"

	"github.com/urfave/cli"
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
func (s *Action) TemplatesPrint(ctx context.Context, c *cli.Context) error {
	tree, err := s.Store.TemplateTree(ctx)
	if err != nil {
		return ExitError(ctx, ExitList, err, "failed to list templates: %s", err)
	}
	fmt.Fprintln(stdout, tree.Format(0))
	return nil
}

// TemplatePrint will lookup and print a single template
func (s *Action) TemplatePrint(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()

	content, err := s.Store.GetTemplate(ctx, name)
	if err != nil {
		return ExitError(ctx, ExitIO, err, "failed to retrieve template: %s", err)
	}

	fmt.Fprintln(stdout, string(content))
	return nil
}

// TemplateEdit will load and existing or new template into an
// editor
func (s *Action) TemplateEdit(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()

	var content []byte
	if s.Store.HasTemplate(ctx, name) {
		var err error
		content, err = s.Store.GetTemplate(ctx, name)
		if err != nil {
			return ExitError(ctx, ExitIO, err, "failed to retrieve template: %s", err)
		}
	} else {
		content = []byte(templateExample)
	}

	ed := editor.Path(c)
	nContent, err := editor.Invoke(ctx, ed, content)
	if err != nil {
		return ExitError(ctx, ExitUnknown, err, "failed to invoke editor %s: %s", ed, err)
	}

	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) {
		return nil
	}

	return s.Store.SetTemplate(ctx, name, nContent)
}

// TemplateRemove will remove a single template
func (s *Action) TemplateRemove(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return ExitError(ctx, ExitUsage, nil, "usage: %s templates remove [name]", s.Name)
	}

	if !s.Store.HasTemplate(ctx, name) {
		return ExitError(ctx, ExitNotFound, nil, "template '%s' not found", name)
	}

	return s.Store.RemoveTemplate(ctx, name)
}

// TemplatesComplete prints a list of all templates for bash completion
func (s *Action) TemplatesComplete(ctx context.Context, c *cli.Context) {
	tree, err := s.Store.TemplateTree(ctx)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return
	}

	for _, v := range tree.List(0) {
		fmt.Fprintln(stdout, v)
	}
}
