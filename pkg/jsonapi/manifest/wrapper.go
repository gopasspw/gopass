package manifest

import (
	"bytes"
	"html/template"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-homedir"
)

const wrapperTemplate = `#!/bin/sh

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

// Render returns the rendered wrapper and manifest
func Render(browser, wrapperPath, binPath string, global bool) ([]byte, []byte, error) {
	mf, err := getManifestContent(browser, wrapperPath)
	if err != nil {
		return nil, nil, err
	}

	if binPath == "" {
		binPath = gopassPath(global)
	}
	wrap, err := getWrapperContent(binPath)
	if err != nil {
		return nil, nil, err
	}

	return wrap, mf, nil
}

func getWrapperContent(gopassPath string) ([]byte, error) {
	tmpl, err := template.New("").Parse(wrapperTemplate)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, struct{ Gopass string }{Gopass: gopassPath}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gopassPath(global bool) string {
	if !global {
		if hd, err := homedir.Dir(); err == nil {
			if gpp, err := os.Executable(); err == nil && strings.HasPrefix(gpp, hd) {
				return gpp
			}
		}
	}
	if gpp, err := exec.LookPath("gopass"); err == nil {
		return gpp
	}
	return "gopass"
}
