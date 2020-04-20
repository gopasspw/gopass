package tests

import (
	"os"
	"path/filepath"
	"runtime"
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

	binName := "gopass"
	if runtime.GOOS == "windows" {
		binName = "gopass.exe"
	}

	bash := `_gopass_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     local IFS=$'\n'
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
 }

complete -F _gopass_bash_autocomplete ` + binName

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
	tp := os.TempDir()
	assert.NoError(t, os.Setenv("PATH", filepath.Join(tp, "foobar")))
	defer func() {
		_ = os.Setenv("PATH", ov)
	}()

	out, err := ts.run("--generate-bash-completion")
	// gopass looks up gpg path in registry. store init  will not fail
	if runtime.GOOS != "windows" {
		assert.Contains(t, out, "Store not initialized")
	}
	assert.NoError(t, err)
}
