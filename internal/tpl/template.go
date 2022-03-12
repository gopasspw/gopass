package tpl

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/gopasspw/gopass/pkg/gopass"
)

type kvstore interface {
	Get(context.Context, string) (gopass.Secret, error)
}

type payload struct {
	Dir     string
	Path    string
	Name    string
	Content string
}

// Execute executes the given template.
func Execute(ctx context.Context, tpl, name string, content []byte, s kvstore) ([]byte, error) {
	funcs := funcMap(ctx, s)
	pl := payload{
		Dir:     filepath.Dir(name),
		Path:    name,
		Name:    filepath.Base(name),
		Content: string(content),
	}

	tmpl, err := template.New(tpl).Funcs(funcs).Parse(tpl)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse template: %w", err)
	}

	buff := &bytes.Buffer{}
	if err := tmpl.Execute(buff, pl); err != nil {
		return []byte{}, fmt.Errorf("failed to execute template: %w", err)
	}

	return buff.Bytes(), nil
}
