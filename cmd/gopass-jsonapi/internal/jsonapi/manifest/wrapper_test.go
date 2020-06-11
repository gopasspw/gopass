package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wrapperGolden = `#!/bin/sh

if [ -f ~/.gpg-agent-info ] && [ -n "$(pgrep gpg-agent)" ]; then
	source ~/.gpg-agent-info
	export GPG_AGENT_INFO
else
	eval $(gpg-agent --daemon)
fi

export PATH="$PATH:/usr/local/bin" # required on MacOS/brew
export GPG_TTY="$(tty)"

gopass-jsonapi listen

exit $?
`
)

func TestWrapperContent(t *testing.T) {
	b, err := getWrapperContent("gopass-jsonapi")
	require.NoError(t, err)
	assert.Equal(t, wrapperGolden, string(b))
}
