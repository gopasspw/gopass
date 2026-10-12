package action

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	fishcomp "github.com/gopasspw/gopass/internal/completion/fish"
	zshcomp "github.com/gopasspw/gopass/internal/completion/zsh"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v3"
)

var escapeRegExp = regexp.MustCompile(`('|"|\s|\(|\)|\<|\>|\&|\;|\#|\:|\\|\||\*|\?)`)

// bashEscape Escape special characters with `\`.
func bashEscape(s string) string {
	return escapeRegExp.ReplaceAllStringFunc(s, func(c string) string {
		if c == `\` {
			return `\\\\`
		}

		if c == `'` {
			return `\` + c
		}

		if c == `"` {
			return `\\\` + c
		}

		return `\\` + c
	})
}

// Complete prints a list of all password names to os.Stdout, for bash completion.
func (s *miscHandler) Complete(ctx context.Context, cmd *cli.Command) {
	ctx = ctxutil.WithGlobalFlags(ctx, cmd)
	_, err := s.Store.IsInitialized(ctx) // important to make sure the structs are not nil.
	if err != nil {
		out.Errorf(ctx, "Store not initialized: %s", err)

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

// CompletionOpenBSDKsh returns an OpenBSD ksh script used for auto completion.
func (s *miscHandler) CompletionOpenBSDKsh(a *cli.Command) error {
	out := `
PASS_LIST=$(gopass ls -f)
set -A complete_gopass -- $PASS_LIST %s
`

	if a == nil {
		return fmt.Errorf("can not parse command options")
	}

	opts := make([]string, 0, len(a.Commands))
	for _, opt := range a.Commands {
		opts = append(opts, opt.Name)
		if len(opt.Aliases) > 0 {
			opts = append(opts, strings.Join(opt.Aliases, " "))
		}
	}

	fmt.Fprintf(stdout, out, strings.Join(opts, " "))

	return nil
}

// CompletionBash returns a bash script used for auto completion.
func (s *miscHandler) CompletionBash(ctx context.Context, cmd *cli.Command) error {
	// This script is based on the bash_autocomplete script shipped with
	// github.com/urfave/cli/v3 (autocomplete/bash_autocomplete). Since the
	// urfave/cli v3 migration the completion output uses a "token:description"
	// format, e.g. "show:Display the content of a secret". Passing these lines
	// to compgen verbatim (as the previous script did) inserts the description
	// into the command line. The script below splits token and description,
	// completes only on the token and offers the description on the second TAB
	// (COMP_TYPE 63).
	//
	// Deviations from the upstream script:
	//
	//  - Secret names completed by the gopass ShellComplete hook can contain
	//    colons. They are emitted with the colon escaped ("websites\:github"),
	//    so the "token:description" split below only happens for the raw
	//    colons urfave/cli emits. compgen unescapes such words again on
	//    insertion.
	//  - The completion output is captured with IFS restricted to newlines, so
	//    secret names containing spaces do not fall apart into several words.
	//  - The completion request is executed directly from the COMP_WORDS array
	//    instead of through eval, and it includes the word being completed when
	//    it starts with a dash, so urfave/cli can filter the flag suggestions
	//    server side.
	//  - No dependency on the bash-completion package: on a bare bash without
	//    _comp_initialize/_get_comp_words_by_ref the function falls back to the
	//    raw COMP_WORDS values.
	out := `__GPASS_INIT() {
  COMPREPLY=()
  if declare -F _comp_initialize >/dev/null 2>&1; then
    _comp_initialize "$@"
  elif declare -F _get_comp_words_by_ref >/dev/null 2>&1; then
    _get_comp_words_by_ref "$@" cur prev words cword
  else
    cur="${COMP_WORDS[COMP_CWORD]}"
    cword="${COMP_CWORD}"
  fi
}

__GPASS_TOKENS=()
__GPASS_DESCRIPTIONS=()

__GPASS_FIND_DESCRIPTION() {
  local candidate="$1"
  local i
  for i in "${!__GPASS_TOKENS[@]}"; do
    if [[ "${__GPASS_TOKENS[i]}" == "${candidate}" ]]; then
      printf '%s' "${__GPASS_DESCRIPTIONS[i]}"
      return 0
    fi
  done
  return 1
}

_GPASS_BASH_AUTOCOMPLETE() {
  local cur opts
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  __GPASS_INIT -n "=:" || return

  # Keep IFS at newline so that candidates containing spaces stay in one word.
  local IFS=$'\n'

  if [[ "${cur}" == -* ]]; then
    opts=$( "${COMP_WORDS[@]:0:$COMP_CWORD}" "${cur}" --generate-shell-completion 2>/dev/null ) || opts=""
  else
    opts=$( "${COMP_WORDS[@]:0:$COMP_CWORD}" --generate-shell-completion 2>/dev/null ) || opts=""
  fi

  # The completion output uses "token:description" lines. Colons in secret
  # names arrive escaped as "\:" and are kept whole; compgen unescapes them
  # again when the match is inserted.
  __GPASS_TOKENS=()
  __GPASS_DESCRIPTIONS=()

  local line token description longest
  longest=0
  while IFS= read -r line; do
    token="${line}"
    description=""
    if [[ "${line}" == *"\:"* ]]; then
      # Escaped colon: a secret name containing a colon. Keep it whole.
      token="${line}"
    elif [[ "${line}" == *:* ]]; then
      token="${line%%:*}"
      description="${line#*:}"
    fi

    if [[ -z "${token}" ]]; then
      continue
    fi

    __GPASS_TOKENS+=("${token}")
    __GPASS_DESCRIPTIONS+=("${description}")
    (( ${#token} > longest )) && longest=${#token}
  done <<< "${opts}"

  local matches=( $(compgen -W "${__GPASS_TOKENS[*]}" -- "${cur}") )

  # COMP_TYPE 63: bash is listing matches (usually the second TAB). Show the
  # descriptions alongside the tokens.
  if [[ "${COMP_TYPE:-}" == "63" && ${#matches[@]} -gt 0 ]]; then
    local listed=()
    local candidate
    for candidate in "${matches[@]}"; do
      local desc
      desc="$(__GPASS_FIND_DESCRIPTION "${candidate}")"
      if [[ -n "${desc}" ]]; then
        local padded
        padded="$(printf '%-*s' "${longest}" "${candidate}")"
        listed+=("${padded}  --  ${desc}")
      else
        listed+=("${candidate}")
      fi
    done
    COMPREPLY=("${listed[@]}")
  else
    COMPREPLY=("${matches[@]}")
  fi

  return 0
}

`
	out = strings.ReplaceAll(out, "__GPASS_INIT", "__"+s.Name+"_init_completion")
	out = strings.ReplaceAll(out, "_GPASS_BASH_AUTOCOMPLETE", "_"+s.Name+"_bash_autocomplete")
	out = strings.ReplaceAll(out, "__GPASS_TOKENS", "__"+s.Name+"_completion_tokens")
	out = strings.ReplaceAll(out, "__GPASS_DESCRIPTIONS", "__"+s.Name+"_completion_descriptions")
	out = strings.ReplaceAll(out, "__GPASS_FIND_DESCRIPTION", "__"+s.Name+"_find_description_for")
	out += "complete -F _" + s.Name + "_bash_autocomplete " + s.Name
	if runtime.GOOS == "windows" {
		out += "\ncomplete -F _" + s.Name + "_bash_autocomplete " + s.Name + ".exe"
	}
	fmt.Fprintln(stdout, out)

	return nil
}

// CompletionFish returns an autocompletion script for fish.
func (s *miscHandler) CompletionFish(a *cli.Command) error {
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

// CompletionZSH returns a zsh completion script.
func (s *miscHandler) CompletionZSH(a *cli.Command) error {
	comp, err := zshcomp.GetCompletion(a)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, comp)

	return nil
}
