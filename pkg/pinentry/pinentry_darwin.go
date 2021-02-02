// +build darwin

package pinentry

import "github.com/gopasspw/gopass/pkg/pinentry/gpgconf"

// GetBinary always returns pinentry-mac
func GetBinary() string {
	if p, err := gpgconf.Path("pinentry"); err == nil && p != "" {
		return p
	}
	return "pinentry-mac"
}
