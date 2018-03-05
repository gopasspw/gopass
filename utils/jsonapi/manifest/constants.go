package manifest

var wrapperTemplate = `#!/bin/sh

if [ -f ~/.gpg-agent-info ] && [ -n "$(pgrep gpg-agent)" ]; then
source ~/.gpg-agent-info
export GPG_AGENT_INFO
else
eval $(gpg-agent --daemon)
fi

export PATH="$PATH:/usr/local/bin" # required on MacOS/brew
export GPG_TTY="$(tty)"
%s jsonapi listen
exit $?`

// DefaultBrowser to select when no browser is specified
var DefaultBrowser = "chrome"

// DefaultWrapperPath where the gopass wrapper shell script is installed to
var DefaultWrapperPath = "/usr/local/bin"

// ValidBrowsers are all browsers for which the manifest can be currently installed
var ValidBrowsers = []string{"chrome", "chromium", "firefox"}

var name = "com.justwatch.gopass"
var wrapperName = "gopass_wrapper.sh"
var description = "Gopass wrapper to search and return passwords"
var connectionType = "stdio"
var chromeOrigins = []string{
	"chrome-extension://kkhfnlkhiapbiehimabddjbimfaijdhk/", // gopassbridge
}
var firefoxOrigins = []string{
	"{eec37db0-22ad-4bf1-9068-5ae08df8c7e9}", // gopassbridge
}

type manifestBase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Type        string `json:"type"`
}

type chromeManifest struct {
	manifestBase
	AllowedOrigins []string `json:"allowed_origins"`
}

type firefoxManifest struct {
	manifestBase
	AllowedExtensions []string `json:"allowed_extensions"`
}
