package action

import (
	"fmt"
	"regexp"

	fishcomp "github.com/justwatchcom/gopass/utils/completion/fish"
	zshcomp "github.com/justwatchcom/gopass/utils/completion/zsh"
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
	out := `_gopass_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     local IFS=$'\n'
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
 }

`
	out += "complete -F _gopass_bash_autocomplete " + s.Name
	fmt.Println(out)

	return nil
}

// CompletionFish returns an autocompletion script for fish
func (s *Action) CompletionFish(c *cli.Context, a *cli.App) error {
	comp, err := fishcomp.GetCompletion(a)
	if err != nil {
		return err
	}

	fmt.Println(comp)
	return nil
}

// CompletionZSH returns a zsh completion script
func (s *Action) CompletionZSH(c *cli.Context, a *cli.App) error {
	comp, err := zshcomp.GetCompletion(a)
	if err != nil {
		return err
	}

	fmt.Println(comp)
	return nil
}
