package tpl

import (
	"bytes"
	"context"
	"path/filepath"
	"text/template"

	"github.com/justwatchcom/gopass/store/secret"
)

type kvstore interface {
	Get(context.Context, string) (*secret.Secret, error)
}

type payload struct {
	Dir     string
	Path    string
	Name    string
	Content string
}

// Execute executes the given template
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
		return []byte{}, err
	}

	buff := &bytes.Buffer{}
	if err := tmpl.Execute(buff, pl); err != nil {
		return []byte{}, err
	}

	return buff.Bytes(), nil
}
