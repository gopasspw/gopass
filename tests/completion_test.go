package tests

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("completion")
	assert.NoError(t, err)
	assert.Contains(t, out, "Source for auto completion in bash")
	assert.Contains(t, out, "Source for auto completion in zsh")

	bash := `_gopass_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     local IFS=$'\n'
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
 }

complete -F _gopass_bash_autocomplete gopass`

	out, err = ts.run("completion bash")
	assert.NoError(t, err)
	assert.Equal(t, bash, out)

	out, err = ts.run("completion zsh")
	assert.NoError(t, err)
	assert.Contains(t, out, "compdef gopass")

	out, err = ts.run("completion fish")
	assert.NoError(t, err)
	assert.Contains(t, out, "complete")
}

func TestCompletionNoPath(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ov := os.Getenv("PATH")
	assert.NoError(t, os.Setenv("PATH", "/tmp/foobar"))
	defer func() {
		_ = os.Setenv("PATH", ov)
	}()

	out, err := ts.run("--generate-bash-completion")
	assert.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Contains(t, out, "Store not initialized")
	}
}
