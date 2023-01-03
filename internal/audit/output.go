package audit

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/set"
	"github.com/gopasspw/gopass/internal/tpl"
	"github.com/gopasspw/gopass/pkg/debug"
)

func (r *Report) PrintResults(ctx context.Context) error {
	if r == nil {
		out.Warning(ctx, "Empty report")

		return nil
	}

	var failed bool
	for _, name := range set.SortedKeys(r.Secrets) {
		s := r.Secrets[name]
		out.Printf(ctx, "%s (age: %s)", name, s.Age.String())
		for _, e := range s.Errors {
			out.Errorf(ctx, "Error: %s", e)

			failed = true
		}
		for _, w := range s.Warnings {
			out.Warningf(ctx, "Warning: %s", w)

			failed = true
		}
	}

	if failed {
		return fmt.Errorf("weak password or duplicates detected")
	}

	return nil
}

func (r *Report) RenderCSV(w io.Writer) error {
	cw := csv.NewWriter(w)

	for _, name := range set.SortedKeys(r.Secrets) {
		if len(r.Secrets[name].Errors) < 1 && len(r.Secrets[name].Warnings) < 1 {
			continue
		}

		if err := cw.Write(r.Secrets[name].Record()); err != nil {
			return err
		}
	}
	cw.Flush()

	return cw.Error()
}

func (r *Report) RenderHTML(w io.Writer) error {
	tplStr := htmlTpl

	if r.Template != "" {
		if buf, err := ioutil.ReadFile(r.Template); err == nil {
			tplStr = string(buf)
		} else {
			debug.Log("failed to load custom template from %s: %s", r.Template, err)
		}
	}

	tmpl, err := template.New("report").Funcs(tpl.PublicFuncMap()).Parse(tplStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(w, r); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

var htmlTpl = `<!DOCTYPE html>
<html lang="en">
  <head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>gopass audit report</title>
</head>
<body>
<table>
  <thead>
  <th>Secret</th>
  <th>Age</th>
  <th>Errors</th>
  <th>Warnings</th>
  </thead>
{{- range .Secrets }}{{ if or .Errors .Warnings }}
  <tr>
    <td>{{ .Name }}</td>
	<td>{{ .Age | roundDuration }}</td>
	<td>{{ .Warnings | join ", " }}</td>
	<td>{{ .Errors | join ", " }}</td>
  </tr>
{{ end }}{{- end }}
</table>
</body>
</html>
`
