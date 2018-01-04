package fish

// see https://fishshell.com/docs/current/commands.html#complete
var fishTemplate = `#!/usr/bin/env fish
{{ $prog := .Name -}}
set PROG '{{ $prog }}'

function __fish_{{ $prog }}_needs_command
  set -l cmd (commandline -opc)
  if [ (count $cmd) -eq 1 -a $cmd[1] = $PROG ]
    return 0
  end
  return 1
end

function __fish_{{ $prog }}_uses_command
  set cmd (commandline -opc)
  if [ (count $cmd) -gt 1 ]
    if [ $argv[1] = $cmd[2] ]
      return 0
    end
  end
  return 1
end

function __fish_{{ $prog }}_print_gpg_keys
  gpg2 --list-keys | grep uid | sed 's/.*<\(.*\)>/\1/'
end

function __fish_{{ $prog }}_print_entries
  eval "{{ $prog }} ls --flat"
  for file in $files
    echo "$file"
  end
end

# erase any existing completions for {{ $prog }}
complete -c $PROG -e
complete -c $PROG -f -n '__fish_{{ $prog }}_needs_command' -a "(__fish_{{ $prog }}_print_entries)"
{{- $gflags := .Flags -}}
{{ range .Commands }}
complete -c $PROG -f -n '__fish_{{ $prog }}_needs_command' -a {{ .Name }} -d 'Command: {{ .Usage }}'
{{- $cmd := .Name -}}
{{- range .Subcommands }}
{{- $subcmd := .Name }}
complete -c $PROG -f -n '__fish_{{ $prog }}_uses_command {{ $cmd }}' -a {{ $subcmd }} -d 'Subcommand: {{ .Usage }}'
{{- if or (eq $cmd "copy") (eq $cmd "move") (eq $cmd "delete") (eq $cmd "show") }}complete -c $PROG -f -n '__fish_{{ $prog }}_uses_command {{ $cmd }}' -a "(__fish_{{ $prog }}_print_entries)"{{ end -}}
{{- range .Flags }}
complete -c $PROG -f -n '__fish_{{ $prog }}_uses_command {{ $cmd }} {{ $subcmd }} {{ if ne (. | formatShortFlag) "" }}-s {{ . | formatShortFlag }} {{ end }}-l {{ . | formatLongFlag }} -d "{{ . | formatFlagUsage }}"'
{{- end }}
{{- range $gflags }}
complete -c $PROG -f -n '__fish_{{ $prog }}_uses_command {{ $cmd }} {{ $subcmd }} {{ if ne (. | formatShortFlag) "" }}-s {{ . | formatShortFlag }} {{ end }}-l {{ . | formatLongFlag }} -d "{{ . | formatFlagUsage }}"'
{{- end }}
{{- end }}
{{- end }}`
