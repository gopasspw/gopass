package manifest

import (
	"bytes"
	"html/template"
)

func getWrapperContent(gopassPath string) ([]byte, error) {
	tmpl, err := template.New("").Parse(wrapperTemplate)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, struct{ Gopass string }{Gopass: gopassPath}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
