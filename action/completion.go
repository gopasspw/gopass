package action

import (
	"fmt"
	"regexp"

	"github.com/urfave/cli"
)

var escapeRegExp = regexp.MustCompile(`(\s|\(|\)|\<|\>|\&|\;|\#|\\|\||\*|\?)`)

// bashEscape Escape special characters with `\`
func bashEscape(s string) string {
	return escapeRegExp.ReplaceAllStringFunc(s, func(c string) string {
		if c == `\` {
			return `\\\\`
		}
		return `\\` + c
	})
}

// Complete prints a list of all password names to os.Stdout
func (s *Action) Complete(*cli.Context) {
	list, err := s.Store.List(0)
	if err != nil {
		return
	}

	for _, v := range list {
		fmt.Println(bashEscape(v))
	}
}

// CompletionBash returns a bash script used for auto completion
func (s *Action) CompletionBash(c *cli.Context) error {
	out := `#!/bin/bash

PROG=gopass

_cli_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     local IFS=$'\n'
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
 }

complete -F _cli_bash_autocomplete $PROG
`
	fmt.Println(out)

	return nil
}

// CompletionZSH returns a script that uses bash's auto completion
func (s *Action) CompletionZSH(c *cli.Context) error {
	out := `autoload -U compinit && compinit
autoload -U bashcompinit && bashcompinit

source <(gopass completion bash)
`
	fmt.Println(out)

	return nil
}
