package action

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
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
# - get "key": e.g. {{ get "path/to/some/other/secret" | md5sum }}
`
)

// TemplatesPrint will pretty-print a tree of templates
func (s *Action) TemplatesPrint(ctx context.Context, c *cli.Context) error {
	tree, err := s.Store.TemplateTree()
	if err != nil {
		return exitError(ctx, ExitList, err, "failed to list templates: %s", err)
	}
	fmt.Fprintln(stdout, tree.Format(0))
	return nil
}

// TemplatePrint will lookup and print a single template
func (s *Action) TemplatePrint(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()

	content, err := s.Store.GetTemplate(ctx, name)
	if err != nil {
		return exitError(ctx, ExitIO, err, "failed to retrieve template: %s", err)
	}

	fmt.Fprintln(stdout, string(content))
	return nil
}

// TemplateEdit will load and existing or new template into an
// editor
func (s *Action) TemplateEdit(ctx context.Context, c *cli.Context) error {
	name := c.Args().First()
	// TODO support editing the root template as well
	if name == "" {
		return errors.Errorf("provide a template name")
	}

	var content []byte
	if s.Store.HasTemplate(ctx, name) {
		var err error
		content, err = s.Store.GetTemplate(ctx, name)
		if err != nil {
			return exitError(ctx, ExitIO, err, "failed to retrieve template: %s", err)
		}
	} else {
		content = []byte(templateExample)
	}

	editor := getEditor(c)
	nContent, err := s.editor(ctx, editor, content)
	if err != nil {
		return exitError(ctx, ExitUnknown, err, "failed to invoke editor %s: %s", editor, err)
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
		return exitError(ctx, ExitUsage, nil, "usage: %s templates remove [name]", s.Name)
	}

	if !s.Store.HasTemplate(ctx, name) {
		return exitError(ctx, ExitNotFound, nil, "template '%s' not found", name)
	}

	return s.Store.RemoveTemplate(ctx, name)
}

// TemplatesComplete prints a list of all templates for bash completion
func (s *Action) TemplatesComplete(*cli.Context) {
	tree, err := s.Store.TemplateTree()
	if err != nil {
		fmt.Fprintln(stdout, err)
		return
	}

	for _, v := range tree.List(0) {
		fmt.Fprintln(stdout, v)
	}
}
