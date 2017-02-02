package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletion(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("completion")
	assert.NoError(t, err)
	assert.Contains(t, out, "Source for auto completion in bash")
	assert.Contains(t, out, "Source for auto completion in zsh")

	bash := `#!/bin/bash

PROG=gopass

_cli_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
 }

complete -F _cli_bash_autocomplete $PROG`

	out, err = ts.run("completion bash")
	assert.NoError(t, err)
	assert.Equal(t, bash, out)

	zsh := `autoload -U compinit && compinit
autoload -U bashcompinit && bashcompinit

source <(gopass completion bash)`

	out, err = ts.run("completion zsh")
	assert.NoError(t, err)
	assert.Equal(t, zsh, out)
}
