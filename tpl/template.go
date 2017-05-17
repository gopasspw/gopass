package tpl

import (
	"bytes"
	"path/filepath"
	"text/template"
)

type kvstore interface {
	Get(string) ([]byte, error)
}

type payload struct {
	Dir     string
	Path    string
	Name    string
	Content string
}

// Execute executes the given template
func Execute(tpl, name string, content []byte, s kvstore) ([]byte, error) {
	funcs := funcMap(s)
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
