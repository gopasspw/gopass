//go:build ignore
// +build ignore

// This program generates pwrules_gen.go. It can be invoked by running
// go generate.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	aliasURL  = "https://raw.githubusercontent.com/apple/password-manager-resources/main/quirks/websites-with-shared-credential-backends.json"
	changeURL = "https://raw.githubusercontent.com/apple/password-manager-resources/main/quirks/change-password-URLs.json"
	rulesURL  = "https://raw.githubusercontent.com/apple/password-manager-resources/main/quirks/password-rules.json"
)

func main() {
	aliases, err := fetchAliases()
	if err != nil {
		panic(err)
	}
	changes, err := fetchChangeURLs()
	if err != nil {
		panic(err)
	}
	rules, err := fetchRules()
	if err != nil {
		panic(err)
	}

	f, err := os.Create("pwrules_gen.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pkgTpl.Execute(f, struct {
		Timestamp time.Time
		URLs      []string
		Aliases   map[string][]string
		Changes   map[string]string
		Rules     map[string]jsonRule
	}{
		Timestamp: time.Now().UTC(),
		URLs: []string{
			aliasURL,
			changeURL,
			rulesURL,
		},
		Aliases: aliases,
		Changes: changes,
		Rules:   rules,
	})
}

func fetchAliases() (map[string][]string, error) {
	resp, err := http.Get(aliasURL)
	if err != nil {
		return nil, err
	}
	var ja [][]string
	if err := json.NewDecoder(resp.Body).Decode(&ja); err != nil {
		return nil, err
	}
	aliases := make(map[string][]string, len(ja))
	for _, as := range ja {
		for _, a := range as {
			aliases[a] = as
		}
	}
	return aliases, nil
}

func fetchChangeURLs() (map[string]string, error) {
	resp, err := http.Get(changeURL)
	if err != nil {
		return nil, err
	}
	var change map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&change); err != nil {
		return nil, err
	}
	return change, nil
}

type jsonRule struct {
	Exact bool   `json:"exact-domain-match-only"`
	Rules string `json:"password-rules"`
}

func fetchRules() (map[string]jsonRule, error) {
	var src io.Reader
	if fn := os.Getenv("PWGEN_RULES_FILE"); fn != "" {
		f, err := os.Open(fn)
		if err != nil {
			return nil, err
		}
		src = f
	} else {
		resp, err := http.Get(rulesURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		src = resp.Body
	}

	var jr map[string]jsonRule
	if err := json.NewDecoder(&cleaningReader{src: src, ign: map[string]int{"launtel.net.au": 5}}).Decode(&jr); err != nil {
		return nil, err
	}
	return jr, nil
}

type cleaningReader struct {
	src io.Reader
	rdr io.Reader
	ign map[string]int // map of domains to ignore, value is number of lines to skip
}

func (c *cleaningReader) init() error {
	if c.rdr != nil {
		return nil
	}
	// no need to do anything if the ignore list is empty
	if len(c.ign) < 1 {
		fmt.Println("ignore list is empty")
		c.rdr = c.src
		return nil
	}

	var buf bytes.Buffer
	scanner := bufio.NewScanner(c.src)
	for scanner.Scan() {
		line := scanner.Text()
		skip := 0
		// skip two broken entries. this is a terrible hack because
		// the JSON is not valid.
		for needle, numSkip := range c.ign {
			want := fmt.Sprintf("\"%s\":", needle)
			if strings.Contains(line, want) {
				fmt.Printf("skipping %d lines after %s\n", numSkip, needle)
				skip = numSkip
			}
		}
		// the broken entries are three lines each. the first was already consumed
		// above, so we need to skip the next two lines to consume all of it.
		for i := 0; i < skip; i++ {
			scanner.Scan()
			fmt.Printf("Skipped line: %s\n", scanner.Text())
		}
		if skip > 0 {
			continue
		}
		buf.WriteString(line)
	}
	c.rdr = bytes.NewReader(buf.Bytes())
	return nil
}

func (c *cleaningReader) Read(p []byte) (n int, err error) {
	if err := c.init(); err != nil {
		return 0, err
	}
	return c.rdr.Read(p)
}

// cf. https://blog.carlmjohnson.net/post/2016-11-27-how-to-use-go-generate/
var pkgTpl = template.Must(template.New("").Parse(`// Code generated by go generate gen.go. DO NOT EDIT.
// This package was generated by go generate gen.go at
// {{ .Timestamp }}
// using data from
// {{- range .URLs }}
// {{ . }}
// {{- end }}
package pwrules

var genAliases = map[string][]string{
{{- range $key, $value := .Aliases }}
  "{{ $key }}": []string{
  {{- range $value }}
    "{{ . }}",
  {{- end }}
  },
{{- end }}
}

var genChange = map[string]string{
{{- range $key, $value := .Changes }}
	"{{ $key }}": "{{ $value }}",
{{- end }}
}

var genRules = map[string]string{
{{- range $key, $value := .Rules }}
	"{{ $key }}": {{ printf "%q" $value.Rules }},
{{- end }}
}

var genRulesExact = map[string]bool{
{{- range $key, $value := .Rules }}
	"{{ $key }}": {{ printf "%t" $value.Exact }},
{{- end }}
}
`))
