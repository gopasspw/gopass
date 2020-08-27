package action

import (
	"fmt"
	"github.com/gopasspw/gopass/internal/tree"
	"regexp"
	"runtime"
	"strings"

	fishcomp "github.com/gopasspw/gopass/internal/completion/fish"
	zshcomp "github.com/gopasspw/gopass/internal/completion/zsh"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
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
func (s *Action) Complete(c *cli.Context) {
	ctx := ctxutil.WithGlobalFlags(c)
	_, err := s.Store.Initialized(ctx) // important to make sure the structs are not nil
	if err != nil {
		out.Error(ctx, "Store not initialized: %s", err)
		return
	}
	list, err := s.Store.List(ctx, tree.INF)
	if err != nil {
		return
	}

	for _, v := range list {
		fmt.Fprintln(stdout, bashEscape(v))
	}
}

// CompletionOpenBSDKsh returns an OpenBSD ksh script used for auto completion
func (s *Action) CompletionOpenBSDKsh(a *cli.App) error {
	out := `
PASS_LIST=$(gopass ls -f)
set -A complete_gopass -- $PASS_LIST %s
`

	if a == nil {
		return fmt.Errorf("can not parse command options")
	}

	var opts []string
	for _, opt := range a.Commands {
		opts = append(opts, opt.Name)
		if len(opt.Aliases) > 0 {
			opts = append(opts, strings.Join(opt.Aliases, " "))
		}
	}

	fmt.Fprintf(stdout, out, strings.Join(opts, " "))
	return nil
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
	if runtime.GOOS == "windows" {
		out += "\ncomplete -F _gopass_bash_autocomplete " + s.Name + ".exe"
	}
	fmt.Fprintln(stdout, out)

	return nil
}

// CompletionFish returns an autocompletion script for fish
func (s *Action) CompletionFish(a *cli.App) error {
	if a == nil {
		return fmt.Errorf("app is nil")
	}
	comp, err := fishcomp.GetCompletion(a)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, comp)
	return nil
}

// CompletionZSH returns a zsh completion script
func (s *Action) CompletionZSH(a *cli.App) error {
	comp, err := zshcomp.GetCompletion(a)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, comp)
	return nil
}
