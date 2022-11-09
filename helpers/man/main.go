// Copyright 2021 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

// Man implements a man(1) documentation generator that is run as part of the
// release helper to generate an up to date manpage for Gopass.
package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
	ap "github.com/gopasspw/gopass/internal/action"
	"github.com/gopasspw/gopass/internal/action/pwgen"
	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/urfave/cli/v2"
)

func main() {
	// this is a workaround for the man helper getting accidentially
	// installed into my $GOBIN dir and me not being able to figure out
	// why. So instead of being greeted with an ugly panic message
	// every now and then when I need to open a man page I decided
	// to rather have a little bit of code to automate this away.
	if len(os.Args) > 0 && os.Args[0] == "man" {
		manPath, err := lookPath("man")
		if err != nil {
			panic(err)
		}
		cmd := exec.Command(manPath, os.Args[1:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			os.Exit(cmd.ProcessState.ExitCode())
		}

		return
	}

	vs, err := ioutil.ReadFile("VERSION")
	if err != nil {
		panic(err)
	}
	version := semver.MustParse(strings.TrimSpace(string(vs)))

	action, err := ap.New(config.New(), version)
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

// from https://cs.opensource.google/go/go/+/refs/tags/go1.17.3:src/os/exec/lp_unix.go
func lookPath(file string) (string, error) {
	curPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := filepath.Join(dir, file)
		// do not call ourselves
		if path == curPath {
			continue
		}
		if err := findExecutable(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("%s: executable file not found in $PATH", file)
}

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0o111 != 0 {
		return nil
	}
	return fs.ErrPermission
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
\fB{{ if (gt (len $alias) 1) }}\-{{ end }}\-{{ $alias }}\fR,{{ end }}
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
\fB{{ if (gt (len $alias) 1) }}\-{{ end }}\-{{ $alias }}\fR,{{ end }}
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
