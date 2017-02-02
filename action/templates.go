package action

import (
	"bytes"
	"fmt"

	"github.com/urfave/cli"
)

// TemplatesPrint will pretty-print a tree of templates
func (s *Action) TemplatesPrint(c *cli.Context) error {
	tree, err := s.Store.TemplateTree()
	if err != nil {
		return err
	}
	fmt.Println(tree.Format())
	return nil
}

// TemplateEdit will load and existing or new template into an
// editor
func (s *Action) TemplateEdit(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("provide a template name")
	}

	var content []byte
	if s.Store.HasTemplate(name) {
		var err error
		content, err = s.Store.GetTemplate(name)
		if err != nil {
			return err
		}
	}

	nContent, err := s.editor(content)
	if err != nil {
		return err
	}

	// If content is equal, nothing changed, exiting
	if bytes.Equal(content, nContent) {
		return nil
	}

	return s.Store.SetTemplate(name, nContent)
}

// TemplateRemove will remove a single template
func (s *Action) TemplateRemove(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("provide a template name")
	}

	if !s.Store.HasTemplate(name) {
		return fmt.Errorf("template not found")
	}

	return s.Store.RemoveTemplate(name)
}

// TemplatesComplete prints a list of all templates for bash completion
func (s *Action) TemplatesComplete(*cli.Context) {
	tree, err := s.Store.TemplateTree()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range tree.List() {
		fmt.Println(v)
	}
	return
}
