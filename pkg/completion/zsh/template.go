package zsh

// see http://zsh.sourceforge.net/Doc/Release/Completion-System.html
var zshTemplate = `{{ $prog := .Name }}#compdef {{ $prog }}

_{{ $prog }} () {
    local cmd
    if (( CURRENT > 2)); then
        cmd=${words[2]}
        curcontext="${curcontext%:*:*}:{{ $prog }}-$cmd"
        (( CURRENT-- ))
        shift words
        case "${cmd}" in
{{- range .Commands }}
          {{ .Name }}{{ range .Aliases }}|{{ . }}{{ end }})
              {{- if .Subcommands }}
              local -a subcommands
              subcommands=({{ range .Subcommands }}
              "{{ .Name }}:{{ .Usage }}"{{ end }}
              )
              {{- end }}
              {{ if .Flags }}_arguments :{{ range .Flags }} "{{ . | formatFlag }}"{{ end }}{{ end }}
              _describe -t commands "{{ $prog }} {{ .Name }}" subcommands
              {{ if or (eq .Name "insert") (eq .Name "generate")  (eq .Name "list") }}_{{ $prog }}_complete_folders{{ end }}
              {{ if or (eq .Name "copy") (eq .Name "move") (eq .Name "delete") (eq .Name "show") (eq .Name "edit") (eq .Name "insert") (eq .Name "generate") }}_{{ $prog }}_complete_passwords{{ end }}
              ;;
{{- end }}
          *)
              _{{ $prog }}_complete_passwords
              ;;
        esac
    else
        local -a subcommands
        subcommands=({{ range .Commands }}
          "{{ .Name }}:{{ .Usage }}"{{ end }}
        )
        _describe -t command '{{ $prog }}' subcommands
        _arguments : {{ range .Flags }}"{{ . | formatFlag }}" {{ end }}
        _{{ $prog }}_complete_passwords
    fi
}

_{{ $prog }}_complete_keys () {
    local IFS=$'\n'
    _values 'gpg keys' $(gpg2 --list-secret-keys --with-colons 2> /dev/null | cut -d : -f 10 | sort -u | sed '/^$/d')
}

_{{ $prog }}_complete_passwords () {
    _arguments : \
        "--clip[Copy the first line of the secret into the clipboard]"
    _values 'passwords' $({{ $prog }} ls --flat)
}

_{{ $prog }}_complete_folders () {
    local -a folders
    folders=("${(@f)$({{ $prog }} ls --folders --flat)}")
    _describe -t folders "folders" folders -qS /
}

_{{ $prog }}`
