package manifest

var (
	wrapperTemplate = `#!/bin/sh

if [ -f ~/.gpg-agent-info ] && [ -n "$(pgrep gpg-agent)" ]; then
	source ~/.gpg-agent-info
	export GPG_AGENT_INFO
else
	eval $(gpg-agent --daemon)
fi

export PATH="$PATH:/usr/local/bin" # required on MacOS/brew
export GPG_TTY="$(tty)"

{{ .Gopass }} jsonapi listen

exit $?
`
)
