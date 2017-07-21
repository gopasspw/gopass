package root

import "github.com/justwatchcom/gopass/gpg"

// GPGVersion returns GPG version information
func (r *Store) GPGVersion() gpg.Version {
	return r.store.GPGVersion()
}
