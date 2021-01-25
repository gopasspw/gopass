// +build !darwin,!windows

package pinentry

import "github.com/gopasspw/gopass/pkg/pinentry/gpgconf"

// GetBinary returns the binary name
func GetBinary() string {
	if p, err := gpgconf.Path("pinentry"); err == nil && p != "" {
		return p
	}
	return "pinentry"
}
