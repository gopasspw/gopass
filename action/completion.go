package action

import (
	"fmt"

	"github.com/urfave/cli"
)

// Complete prints a list of all password names to os.Stdout
func (s *Action) Complete(*cli.Context) {
	list, err := s.Store.List()
	if err != nil {
		return
	}

	for _, v := range list {
		fmt.Println(v)
	}
}

// CompletionBash returns a bash script used for auto completion
func (s *Action) CompletionBash(c *cli.Context) error {
	out := `#!/bin/bash

PROG=gopass

_cli_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
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

// CompletionDMenu returns a script that starts dmenu
// Usage: eval "$(gopass completion dmenu)"
func (s *Action) CompletionDMenu(c *cli.Context) error {
	out := `#!/usr/bin/env bash

shopt -s nullglob globstar

typeit=0
if [[ $1 == "--type" ]]; then
	typeit=1
	shift
fi

password=$(gopass list --raw | dmenu "$@")

[[ -n $password ]] || exit

if [[ $typeit -eq 0 ]]; then
	gopass show -c "$password" 2>/dev/null
else
	gopass show "$password" | { read -r pass; printf %s "$pass"; } |
		xdotool type --clearmodifiers --file -
fi
`
	fmt.Println(out)

	return nil
}
