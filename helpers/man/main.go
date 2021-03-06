package main

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
	ap "github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/action/pwgen"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/urfave/cli/v2"
)

func main() {
	vs, err := ioutil.ReadFile("VERSION")
	if err != nil {
		panic(err)
	}
	version := semver.MustParse(strings.TrimSpace(string(vs)))

	action, err := ap.New(&config.Config{}, version)
	if err != nil {
		panic(err)
	}

	cmds := action.GetCommands()
	cmds = append(cmds, pwgen.GetCommands()...)
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })

	data := &payload{
		SectionNumber: 1,
		DatePretty:    time.Now().UTC().Format("January 2006"),
		Version:       version.String(),
		SectionName:   "User Commands",
		Commands:      cmds,
		Flags:         getFlags(ap.ShowFlags()),
	}
	funcMap := template.FuncMap{
		"flags": getFlags,
	}
	if err := template.Must(template.New("man").Funcs(funcMap).Parse(manTpl)).Execute(os.Stdout, data); err != nil {
		panic(err)
	}
}

func getFlags(flags []cli.Flag) []flag {
	sort.Slice(flags, func(i, j int) bool { return flags[i].Names()[0] < flags[j].Names()[0] })

	out := make([]flag, 0, len(flags))
	for _, f := range flags {
		switch v := f.(type) {
		case *cli.BoolFlag:
			out = append(out, flag{
				Name:        v.Name,
				Aliases:     append([]string{v.Name}, v.Aliases...),
				Description: v.Usage,
			})
		case *cli.IntFlag:
			out = append(out, flag{
				Name:        v.Name,
				Aliases:     append([]string{v.Name}, v.Aliases...),
				Description: v.Usage,
			})
		case *cli.StringFlag:
			out = append(out, flag{
				Name:        v.Name,
				Aliases:     append([]string{v.Name}, v.Aliases...),
				Description: v.Usage,
			})
		}
	}
	return out
}

type flag struct {
	Name        string
	Aliases     []string
	Description string
}

type payload struct {
	SectionNumber int    // 1
	DatePretty    string // July 2020
	Version       string // 1.12.1
	SectionName   string // User Commands
	Commands      []*cli.Command
	Flags         []flag
}

var manTpl = `
.TH GOPASS "{{ .SectionNumber }}" "{{ .DatePretty }}" "gopass (github.com/gopasspw/gopass) {{ .Version }}" "{{ .SectionName }}"
.SH NAME
gopass - The standard Unix password manager
.SH SYNOPSIS
.B gopass
[\fI\,global options\/\fR] \fI\,command\/\fR [\fI\,command options\/\fR] [\fI,arguments\/\fR...]
.SH GLOBAL OPTIONS
{{ range $flag := .Flags }}
.TP{{ range $alias := $flag.Aliases }}
\fB\-\-{{ $alias }}\fR,{{ end }}
{{ $flag.Description }}{{ end }}
.SH COMMANDS
{{ range $cmd := .Commands }}
.SS {{ $cmd.Name }}
{{ $cmd.Usage }}

{{ $cmd.Description }}
{{- if $cmd.Flags }}

.B Flags
{{- range $flag := $cmd.Flags | flags }}
.TP{{ range $alias := $flag.Aliases }}
\fB\-\-{{ $alias }}\fR,{{ end }}
{{ $flag.Description }}{{ end }}
{{- end }}
{{- end}}

.SH "REPORTING BUGS"
Report bugs to <https://github.com/gopasspw/gopass/issues/new>
.SH "COPYRIGHT"
Copyright \(co 2021 Gopass Authors
This program is free software; you may redistribute it under the terms of
the MIT license. This program has absolutely no warranty.
`
