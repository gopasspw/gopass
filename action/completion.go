package action

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/urfave/cli"
)

// Complete prints a list of all password names to os.Stdout
func (s *Action) Complete(*cli.Context) {
	list, err := s.Store.List(0)
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
	typeit := c.Bool("type")
	args := c.String("args")

	list, err := s.Store.List(0)
	if err != nil {
		return err
	}

	argsSplit, err := shellquote.Split(args)
	if err != nil {
		return err
	}

	name, err := dmenu(list, argsSplit...)
	if err != nil {
		return err
	}

	content, err := s.Store.First(name)
	if err != nil {
		return err
	}

	if typeit {
		return exec.Command("xdotool", "type", "--clearmodifiers", "--", string(content)).Run()
	}

	return s.copyToClipboard(name, content)
}

// dmenu runs it with the provided strings and returns the selected string
func dmenu(list []string, args ...string) (string, error) {
	stdin := bytes.NewBuffer(nil)
	for _, v := range list {
		_, _ = stdin.WriteString(v + "\n")
	}

	cmd := exec.Command("dmenu", args...)
	cmd.Stdin = stdin
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
