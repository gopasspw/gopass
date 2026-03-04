package config

import "github.com/gopasspw/gitconfig"

func init() {
	// Disable unescaping of values. This is not strictly conformant with the
	// git config spec, but it avoid interpreting backslashes in paths as linebreaks
	// or tabs.
	gitconfig.CompatMode = true
}
