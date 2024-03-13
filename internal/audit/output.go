package audit

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

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

	debug.Log("Printing results for %d secrets", len(r.Secrets))

	var failed bool
	for _, name := range set.SortedKeys(r.Secrets) {
		s := r.Secrets[name]
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s (age: %s) ", name, s.HumanizeAge()))
		if !s.HasFindings() {
			sb.WriteString("OK")
			out.OK(ctx, sb.String())

			continue
		}
		sb.WriteString("Potentially weak. ")
		for k, v := range s.Findings {
			if v.Severity == "error" || v.Severity == "warning" {
				failed = true
			}

			switch v.Severity {
			case "error":
				fallthrough
			case "warning":
				sb.WriteString(fmt.Sprintf("%s: %s. ", k, v.Message))
			default:
				continue
			}
		}
		out.Warning(ctx, sb.String())
	}

	if failed {
		return fmt.Errorf("weak password or duplicates detected")
	}

	return nil
}

func (r *Report) PrintSummary(ctx context.Context) error {
	if r == nil {
		out.Warning(ctx, "Empty report")

		return nil
	}

	debug.Log("Printing summary for %d findings", len(r.Findings))

	for _, name := range set.SortedKeys(r.Findings) {
		f := r.Findings[name]
		if f.Len() < 1 {
			continue
		}
		// TODO add details about the analyzer, not just the name
		out.Printf(ctx, "Analyzer %s found issues: ", name)
		for _, v := range set.Sorted(f.Elements()) {
			out.Printf(ctx, "- %s", v)
		}
	}

	if len(r.Findings) > 0 {
		return fmt.Errorf("weak password or duplicates detected")
	}

	return nil
}

func (r *Report) RenderCSV(w io.Writer) error {
	cw := csv.NewWriter(w)

	cs := set.New[string]()
	for _, v := range r.Secrets {
		for k := range v.Findings {
			cs.Add(k)
		}
	}
	cats := cs.Elements()
	sort.Strings(cats)

	for _, name := range set.SortedKeys(r.Secrets) {
		sec := r.Secrets[name]

		rec := make([]string, 0, len(cats)+2)
		rec = append(rec, name)
		rec = append(rec, sec.Age.String())
		for _, cat := range cats {
			if f, found := sec.Findings[cat]; found {
				rec = append(rec, f.Message)

				continue
			}

			rec = append(rec, "ok")
		}

		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()

	return cw.Error()
}

func (r *Report) RenderHTML(w io.Writer) error {
	tplStr := htmlTpl

	if r.Template != "" {
		if buf, err := os.ReadFile(r.Template); err == nil {
			tplStr = string(buf)
		} else {
			debug.Log("failed to load custom template from %s: %s", r.Template, err)
		}
	}

	tmpl, err := template.New("report").Funcs(tpl.PublicFuncMap()).Parse(tplStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(w, getHTMLPayload(r)); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func getHTMLPayload(r *Report) *htmlPayload {
	h := &htmlPayload{
		Today:      time.Now().UTC(),
		Num:        len(r.Secrets),
		Duration:   r.Duration,
		Categories: make([]string, 0, 24),
		Secrets:    make(map[string]SecretReport, len(r.Secrets)),
	}

	cs := set.New[string]()
	for _, v := range r.Secrets {
		for k := range v.Findings {
			cs.Add(k)
		}
	}
	h.Categories = cs.Elements()
	sort.Strings(h.Categories)

	for k, v := range r.Secrets {
		sr := SecretReport{
			Name:     v.Name,
			Age:      v.Age,
			Findings: make(map[string]Finding, len(v.Findings)),
		}
		for _, cat := range h.Categories {
			if f, found := v.Findings[cat]; found {
				sr.Findings[cat] = f

				continue
			}

			sr.Findings[cat] = Finding{
				Severity: "none",
				Message:  "ok",
			}
		}
		h.Secrets[k] = sr
	}

	return h
}

type htmlPayload struct {
	Today      time.Time
	Num        int
	Duration   time.Duration
	Categories []string
	Secrets    map[string]SecretReport
}

var htmlTpl = `<!DOCTYPE html>
<html lang="en">
  <head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>gopass audit report generated on {{ .Today | date }}</title>
  <style>
#findings {
  font-family: Arial, Helvetica, sans-serif;
  border-collapse: collapse;
  width: 100%;
}
#findings td, #findings th {
  border: 1px solid #ddd;
  padding: 8px;
}
#findings tr:nth-child(even){
  background-color: #f3f3f3;
}
#findings tr:hover {
  background-color: #ddd;
}
#findings th {
  padding-top: 12px;
  padding-bottom: 12px;
  text-align: left;
  background-color: #03995D;
  color: white;
}
  </style>
</head>
<body>

Audited {{ .Num }} secrets in {{ .Duration | roundDuration }} on {{ .Today | date }}.<br />

<table id="findings">
  <thead>
  <th>Secret</th>
{{ $cats := .Categories}}
{{- range .Categories }}
<th>{{ . }}</th>
{{ end }}
  </thead>
{{- range .Secrets }}
  <tr>
    <td>{{ .Name }}</td>
{{- range .Findings }}
    <td class="{{ .Severity }}">
        <div title="{{ .Message }}">{{ .Message | truncate 120 }}</div>
    </td>
{{- end }}
  </tr>
{{- end }}
</table>
</body>
</html>
`
